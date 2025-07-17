package cli

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/AlexsanderHamir/prof/analyzer"
	"github.com/AlexsanderHamir/prof/benchmark"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/spf13/cobra"
)

var (
	ErremptyBenchmarks = errors.New("benchmarks argument cannot be an empty list")
	ErremptyProfiles   = errors.New("profiles argument cannot be an empty list")
	Errbracket         = errors.New("argument must be wrapped in brackets")
)

// Arguments holds the CLI arguments for the prof tool.
type Arguments struct {
	Version        bool
	Command        string
	CreateTemplate bool
	OutputPath     string
	Benchmarks     string
	Profiles       string
	Tag            string
	Count          int

	// Performs analyzes on specified profile, according to specified configuration
	// and saves the results in a different file under the AI directory.
	GeneralAnalyze bool

	// Rewrites the profile file instead of saving an analysis in a different place,
	// useful for flagging requests.
	FlagProfiles bool
}

// ParseArguments parses CLI arguments using cobra and returns an Arguments struct.
func ParseArguments() (*Arguments, error) {
	var args Arguments

	var rootCmd = &cobra.Command{
		Use:   "prof",
		Short: "CLI tool for organizing and analyzing Go benchmarks with AI",
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&args.Version, "version", "v", false, "Show version information")
	rootCmd.Flags().StringVar(&args.Benchmarks, "benchmarks", "", "Benchmarks to run (e.g., '[BenchmarkGenPool,BenchmarkSyncPool]')")
	rootCmd.Flags().StringVar(&args.Profiles, "profiles", "", "Profiles to use (e.g., '[cpu,memory,mutex]')")
	rootCmd.Flags().StringVar(&args.Tag, "tag", "", "Tag for the run")
	rootCmd.Flags().IntVar(&args.Count, "count", 0, "Number of runs")
	rootCmd.Flags().BoolVar(&args.GeneralAnalyze, "general_analyze", false, "Run general AI analysis")
	rootCmd.Flags().BoolVar(&args.FlagProfiles, "flag_profiles", false, "Flag profiles for review")

	var setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Set up configuration for the benchmarking tool",
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			args.Command = "setup"
			return nil
		},
	}

	setupCmd.Flags().BoolVar(&args.CreateTemplate, "create-template", false, "Generate a new template configuration file")
	setupCmd.Flags().StringVar(&args.OutputPath, "output-path", "./config_template.json", "Destination path for the template")

	rootCmd.AddCommand(setupCmd)

	if err := rootCmd.Execute(); err != nil {
		return nil, err
	}

	return &args, nil
}

// ValidateRequiredArgs checks if the required arguments are present for running the main command.
func ValidateRequiredArgs(args *Arguments) bool {
	if args.Command != "" || args.Version {
		return true
	}
	return args.Benchmarks != "" && args.Profiles != "" && args.Tag != "" && args.Count > 0
}

// ParseBenchmarkConfig parses the benchmarks and profiles arguments into string slices.
func ParseBenchmarkConfig(benchmarks, profiles string) ([]string, []string, error) {
	if err := validateListArguments(benchmarks, profiles); err != nil {
		return nil, nil, err
	}

	benchmarkList := parseListArgument(benchmarks)
	profileList := parseListArgument(profiles)

	return benchmarkList, profileList, nil
}

// validateListArguments checks if the benchmarks and profiles arguments are valid lists.
func validateListArguments(benchmarks, profiles string) error {
	if strings.TrimSpace(benchmarks) == "[]" {
		return ErremptyBenchmarks
	}
	if strings.TrimSpace(profiles) == "[]" {
		return ErremptyProfiles
	}

	benchmarks = strings.TrimSpace(benchmarks)
	profiles = strings.TrimSpace(profiles)

	if !strings.HasPrefix(benchmarks, "[") || !strings.HasSuffix(benchmarks, "]") {
		return fmt.Errorf("benchmarks %w %s", Errbracket, benchmarks)
	}
	if !strings.HasPrefix(profiles, "[") || !strings.HasSuffix(profiles, "]") {
		return fmt.Errorf("profiles %w %s", Errbracket, profiles)
	}

	return nil
}

// parseListArgument parses a bracketed, comma-separated string into a slice of strings.
func parseListArgument(arg string) []string {
	arg = strings.Trim(arg, "[]")
	if arg == "" {
		return []string{}
	}

	parts := strings.Split(arg, ",")
	var result []string
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

// SetupDirectories delegates directory setup to the benchmark package.
func SetupDirectories(tag string, benchmarks, profiles []string) error {
	return benchmark.SetupDirectories(tag, benchmarks, profiles)
}

// PrintConfiguration prints the parsed configuration and benchmark filter details.
func PrintConfiguration(benchmarks, profiles []string, tag string, count int, benchmarkConfigs map[string]config.BenchmarkFilter) {
	log.Printf("\nParsed arguments:\n")
	log.Printf("Benchmarks: %v\n", benchmarks)
	log.Printf("Profiles: %v\n", profiles)
	log.Printf("Tag: %s\n", tag)
	log.Printf("Count: %d\n", count)

	if len(benchmarkConfigs) > 0 {
		log.Printf("\nBenchmark Function Filter Configurations:\n")
		for benchmark, cfg := range benchmarkConfigs {
			log.Printf("  %s:\n", benchmark)
			log.Printf("    Prefixes: %v\n", cfg.Prefixes)
			if cfg.Ignore != "" {
				log.Printf("    Ignore: %s\n", cfg.Ignore)
			}
		}
	} else {
		log.Printf("\nNo benchmark configuration found in config file - analyzing all functions\n")
	}
}

// RunBenchmarksAndProcessProfiles runs the full benchmark pipeline for each benchmark.
func RunBenchmarksAndProcessProfiles(benchmarks, profiles []string, count int, tag string, benchmarkConfigs map[string]config.BenchmarkFilter) error {
	log.Printf("\nStarting benchmark pipeline...\n")

	for _, benchmarkName := range benchmarks {
		log.Printf("\nRunning benchmark: %s\n", benchmarkName)
		if err := benchmark.RunBenchmark(benchmarkName, profiles, count, tag); err != nil {
			return fmt.Errorf("failed to run benchmark %s: %w", benchmarkName, err)
		}

		log.Printf("\nProcessing profiles for %s...\n", benchmarkName)
		if err := benchmark.ProcessProfiles(benchmarkName, profiles, tag); err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		log.Printf("\nAnalyzing profile functions for %s...\n", benchmarkName)
		if err := benchmark.CollectProfileFunctions(tag, profiles, benchmarkName, benchmarkConfigs[benchmarkName]); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		log.Printf("Completed pipeline for benchmark: %s\n", benchmarkName)
	}

	log.Printf("\nAll benchmarks and profile processing completed successfully!\n")
	return nil
}

// AnalyzeProfiles runs AI analysis for the given tag and profiles using the provided config.
func AnalyzeProfiles(tag string, profiles []string, cfg *config.Config, isFlag bool) error {
	var benchmarks []string
	var profileTypes []string

	if cfg.AIConfig.AllBenchmarks {
		var err error
		benchmarks, err = analyzer.ValidateBenchmarkDirectories(tag)
		if err != nil {
			return err
		}
	} else {
		benchmarks = cfg.AIConfig.SpecificBenchmarks
	}

	if cfg.AIConfig.AllProfiles {
		profileTypes = profiles
	} else {
		profileTypes = cfg.AIConfig.SpecificProfiles
	}

	log.Printf("Found %v benchmarks and %v profile types\n", benchmarks, profileTypes)

	return analyzer.AnalyzeAllProfiles(tag, benchmarks, profileTypes, cfg, isFlag)
}

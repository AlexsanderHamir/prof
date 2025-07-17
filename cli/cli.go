package cli

import (
	"fmt"
	"strings"

	"github.com/AlexsanderHamir/prof/analyzer"
	"github.com/AlexsanderHamir/prof/benchmark"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/spf13/cobra"
)

type Arguments struct {
	Version        bool
	Command        string
	CreateTemplate bool
	OutputPath     string
	Benchmarks     string
	Profiles       string
	Tag            string
	Count          int
	GeneralAnalyze bool
	FlagProfiles   bool
}

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

func ValidateRequiredArgs(args *Arguments) bool {
	if args.Command != "" || args.Version {
		return true
	}
	return args.Benchmarks != "" && args.Profiles != "" && args.Tag != "" && args.Count > 0
}

func ParseBenchmarkConfig(benchmarks, profiles string) ([]string, []string, error) {
	if err := validateListArguments(benchmarks, profiles); err != nil {
		return nil, nil, err
	}

	benchmarkList := parseListArgument(benchmarks)
	profileList := parseListArgument(profiles)

	return benchmarkList, profileList, nil
}

func validateListArguments(benchmarks, profiles string) error {
	if strings.TrimSpace(benchmarks) == "[]" {
		return fmt.Errorf("benchmarks argument cannot be an empty list")
	}
	if strings.TrimSpace(profiles) == "[]" {
		return fmt.Errorf("profiles argument cannot be an empty list")
	}

	benchmarks = strings.TrimSpace(benchmarks)
	profiles = strings.TrimSpace(profiles)

	if !strings.HasPrefix(benchmarks, "[") || !strings.HasSuffix(benchmarks, "]") {
		return fmt.Errorf("benchmarks argument must be wrapped in brackets")
	}
	if !strings.HasPrefix(profiles, "[") || !strings.HasSuffix(profiles, "]") {
		return fmt.Errorf("profiles argument must be wrapped in brackets")
	}

	return nil
}

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

func SetupDirectories(tag string, benchmarks, profiles []string) error {
	return benchmark.SetupDirectories(tag, benchmarks, profiles)
}

func PrintConfiguration(benchmarks, profiles []string, tag string, count int, benchmarkConfigs map[string]config.BenchmarkFilter) {
	fmt.Printf("\nParsed arguments:\n")
	fmt.Printf("Benchmarks: %v\n", benchmarks)
	fmt.Printf("Profiles: %v\n", profiles)
	fmt.Printf("Tag: %s\n", tag)
	fmt.Printf("Count: %d\n", count)

	if len(benchmarkConfigs) > 0 {
		fmt.Printf("\nBenchmark Function Filter Configurations:\n")
		for benchmark, cfg := range benchmarkConfigs {
			fmt.Printf("  %s:\n", benchmark)
			fmt.Printf("    Prefixes: %v\n", cfg.Prefixes)
			if cfg.Ignore != "" {
				fmt.Printf("    Ignore: %s\n", cfg.Ignore)
			}
		}
	} else {
		fmt.Printf("\nNo benchmark configuration found in config file - analyzing all functions\n")
	}
}

func RunBenchmarksAndProcessProfiles(benchmarks, profiles []string, count int, tag string, benchmarkConfigs map[string]config.BenchmarkFilter) error {
	fmt.Printf("\nStarting benchmark pipeline...\n")

	for _, benchmarkName := range benchmarks {
		fmt.Printf("\nRunning benchmark: %s\n", benchmarkName)
		if err := benchmark.RunBenchmark(benchmarkName, profiles, count, tag); err != nil {
			return fmt.Errorf("failed to run benchmark %s: %w", benchmarkName, err)
		}

		fmt.Printf("\nProcessing profiles for %s...\n", benchmarkName)
		if err := benchmark.ProcessProfiles(benchmarkName, profiles, tag); err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		fmt.Printf("\nAnalyzing profile functions for %s...\n", benchmarkName)
		if err := benchmark.AnalyzeProfileFunctions(tag, profiles, benchmarkName, benchmarkConfigs[benchmarkName]); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		fmt.Printf("Completed pipeline for benchmark: %s\n", benchmarkName)
	}

	fmt.Printf("\nAll benchmarks and profile processing completed successfully!\n")
	return nil
}

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

	fmt.Printf("Found %v benchmarks and %v profile types\n", benchmarks, profileTypes)

	return analyzer.AnalyzeAllProfiles(tag, benchmarks, profileTypes, cfg, isFlag)
}

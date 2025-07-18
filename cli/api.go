package cli

import (
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/prof/analyzer"
	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/benchmark"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/spf13/cobra"
)

// ParseArguments parses CLI arguments using cobra and returns an Arguments struct.
func ParseArguments() (*Arguments, error) {
	var args Arguments

	var rootCmd = &cobra.Command{
		Use:   "prof",
		Short: "CLI tool for organizing and analyzing Go benchmarks with AI",
		RunE: func(_ *cobra.Command, _ []string) error {
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
		RunE: func(_ *cobra.Command, _ []string) error {
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
	hasCommandOrVersion := args.Command != "" || args.Version
	if hasCommandOrVersion {
		return true
	}

	hasBenchmarkArgs := args.Benchmarks != "" && args.Profiles != "" && args.Tag != "" && args.Count > 0
	return hasBenchmarkArgs
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

// SetupDirectories delegates directory setup to the benchmark package.
func SetupDirectories(tag string, benchmarks, profiles []string) error {
	return benchmark.SetupDirectories(tag, benchmarks, profiles)
}

// PrintConfiguration prints the parsed configuration and benchmark filter details.
func PrintConfiguration(benchmarks, profiles []string, tag string, count int, functionFilterPerBench map[string]config.FunctionCollectionFilter) {
	slog.Info("Parsed arguments", "Benchmarks", benchmarks, "Profiles", profiles, "Tag", tag, "Count", count)

	hasBenchFunctionFilters := len(functionFilterPerBench) > 0
	if hasBenchFunctionFilters {
		slog.Info("Benchmark Function Filter Configurations:")
		for benchmark, cfg := range functionFilterPerBench {
			slog.Info("Benchmark Config", "Benchmark", benchmark, "Prefixes", cfg.IncludePrefixes, "Ignore", cfg.IgnoreFunctions)
		}
	} else {
		slog.Info("No benchmark configuration found in config file - analyzing all functions")
	}
}

// RunBenchmarksAndProcessProfiles runs the full benchmark pipeline for each benchmark.
func RunBenchmarksAndProcessProfiles(benchmarks, profiles []string, count int, tag string, benchmarkConfigs map[string]config.FunctionCollectionFilter) error {
	slog.Info("Starting benchmark pipeline...")

	for _, benchmarkName := range benchmarks {
		slog.Info("Running benchmark", "Benchmark", benchmarkName)
		if err := benchmark.RunBenchmark(benchmarkName, profiles, count, tag); err != nil {
			return fmt.Errorf("failed to run benchmark %s: %w", benchmarkName, err)
		}

		slog.Info("Processing profiles", "Benchmark", benchmarkName)
		if err := benchmark.ProcessProfiles(benchmarkName, profiles, tag); err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		slog.Info("Analyzing profile functions", "Benchmark", benchmarkName)

		args := &args.CollectionArgs{
			Tag:             tag,
			Profiles:        profiles,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: benchmarkConfigs[benchmarkName],
		}

		if err := benchmark.CollectProfileFunctions(args); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info("All benchmarks and profile processing completed successfully!")
	return nil
}

// AnalyzeProfiles runs AI analysis for the given tag and profiles using the provided config.
func AnalyzeProfiles(tag string, profiles []string, cfg *config.Config, isFlagging bool) error {
	var benchmarks []string
	var profileTypes []string

	if cfg.AIConfig.AllBenchmarks {
		var err error
		benchmarks, err = analyzer.ValidateBenchmarkDirectories(tag, nil)
		if err != nil {
			return err
		}
	} else {
		var err error
		benchmarks, err = analyzer.ValidateBenchmarkDirectories(tag, cfg.AIConfig.SpecificBenchmarks)
		if err != nil {
			return err
		}
	}

	if cfg.AIConfig.AllProfiles {
		profileTypes = profiles
	} else {
		profileTypes = cfg.AIConfig.SpecificProfiles
	}

	slog.Info("Found benchmarks and profile types", "Benchmarks", benchmarks, "ProfileTypes", profileTypes)

	return analyzer.AnalyzeAllProfiles(tag, benchmarks, profileTypes, cfg, isFlagging)
}

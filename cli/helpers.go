package cli

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlexsanderHamir/prof/analyzer"
	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/benchmark"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/shared"
	"github.com/AlexsanderHamir/prof/tracker"
)

var validProfiles = map[string]bool{
	"cpu":    true,
	"memory": true,
	"mutex":  true,
	"block":  true,
}

// validateListArguments checks if the benchmarks and profiles arguments are valid lists.
func validateListArguments(benchmarks, profiles string) error {
	if strings.TrimSpace(benchmarks) == "[]" {
		return ErrEmptyBenchmarks
	}
	if strings.TrimSpace(profiles) == "[]" {
		return ErrEmptyProfiles
	}

	benchmarks = strings.TrimSpace(benchmarks)
	profiles = strings.TrimSpace(profiles)

	if !strings.HasPrefix(benchmarks, "[") || !strings.HasSuffix(benchmarks, "]") {
		return fmt.Errorf("benchmarks %w %s", ErrBracket, benchmarks)
	}
	if !strings.HasPrefix(profiles, "[") || !strings.HasSuffix(profiles, "]") {
		return fmt.Errorf("profiles %w %s", ErrBracket, profiles)
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

// Rest of the functions remain the same...
func parseBenchmarkConfig(benchmarks, profiles string) ([]string, []string, error) {
	if err := validateListArguments(benchmarks, profiles); err != nil {
		return nil, nil, err
	}

	benchmarkList := parseListArgument(benchmarks)
	profileList := parseListArgument(profiles)

	if err := validateAcceptedProfiles(profileList); err != nil {
		return nil, nil, err
	}

	return benchmarkList, profileList, nil
}

func validateAcceptedProfiles(profiles []string) error {
	for _, profile := range profiles {
		if valid := validProfiles[profile]; !valid {
			return fmt.Errorf("received unvalid profile :%s", profile)
		}
	}
	return nil
}

func setupDirectories(tag string, benchmarks, profiles []string) error {
	return benchmark.SetupDirectories(tag, benchmarks, profiles)
}

func printConfiguration(benchArgs *args.BenchArgs, functionFilterPerBench map[string]config.FunctionFilter) {
	slog.Info("Parsed arguments",
		"Benchmarks", benchArgs.Benchmarks,
		"Profiles", benchArgs.Profiles,
		"Tag", benchArgs.Tag,
		"Count", benchArgs.Count)

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

func runBencAndGetProfiles(benchArgs *args.BenchArgs, benchmarkConfigs map[string]config.FunctionFilter) error {
	slog.Info("Starting benchmark pipeline...")

	for _, benchmarkName := range benchArgs.Benchmarks {
		slog.Info("Running benchmark", "Benchmark", benchmarkName)
		if err := benchmark.RunBenchmark(benchmarkName, benchArgs.Profiles, benchArgs.Count, benchArgs.Tag); err != nil {
			return fmt.Errorf("failed to run %s: %w", benchmarkName, err)
		}

		slog.Info("Processing profiles", "Benchmark", benchmarkName)
		if err := benchmark.ProcessProfiles(benchmarkName, benchArgs.Profiles, benchArgs.Tag); err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		slog.Info("Analyzing profile functions", "Benchmark", benchmarkName)

		args := &args.CollectionArgs{
			Tag:             benchArgs.Tag,
			Profiles:        benchArgs.Profiles,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: benchmarkConfigs[benchmarkName],
		}

		if err := benchmark.CollectProfileFunctions(args); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info(shared.InfoCollectionSuccess)
	return nil
}

func analyzeProfiles(tag string, profiles []string, cfg *config.Config, isFlagging bool) error {
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

func printSummary(report *tracker.ProfileChangeReport) {
	fmt.Println("\n=== Performance Tracking Summary ===")

	var regressions, improvements, stable int

	for _, change := range report.FunctionChanges {
		switch change.ChangeType {
		case "REGRESSION":
			regressions++
		case "IMPROVEMENT":
			improvements++
		default:
			stable++
		}
	}

	fmt.Printf("Total Functions Analyzed: %d\n", len(report.FunctionChanges))
	fmt.Printf("Regressions: %d\n", regressions)
	fmt.Printf("Improvements: %d\n", improvements)
	fmt.Printf("Stable: %d\n", stable)

	if regressions > 0 {
		fmt.Println("\n⚠️  Top Regressions:")
		for _, change := range report.FunctionChanges {
			if change.ChangeType == "REGRESSION" {
				fmt.Printf("  • %s\n", change.Summary())
			}
		}
	}

	if improvements > 0 {
		fmt.Println("\n✅ Top Improvements:")
		for _, change := range report.FunctionChanges {
			if change.ChangeType == "IMPROVEMENT" {
				fmt.Printf("  • %s\n", change.Summary())
			}
		}
	}
}

func printDetailedReport(report *tracker.ProfileChangeReport) {
	for i, change := range report.FunctionChanges {
		if i > 0 {
			fmt.Println() // Add spacing between reports
		}
		fmt.Print(change.Report())
	}
}

package cli

import (
	"fmt"
	"log/slog"
	"math"
	"sort"
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
			return fmt.Errorf("received unvalid profile: %s", profile)
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

	var regressionList, improvementList []*tracker.FunctionChangeResult
	var stable int

	// Separate changes by type
	for _, change := range report.FunctionChanges {
		switch change.ChangeType {
		case "REGRESSION":
			regressionList = append(regressionList, change)
		case "IMPROVEMENT":
			improvementList = append(improvementList, change)
		default:
			stable++
		}
	}

	// Sort regressions by percentage (biggest regression first)
	sort.Slice(regressionList, func(i, j int) bool {
		return regressionList[i].FlatChangePercent > regressionList[j].FlatChangePercent
	})

	// Sort improvements by absolute percentage (biggest improvement first)
	sort.Slice(improvementList, func(i, j int) bool {
		return math.Abs(improvementList[i].FlatChangePercent) > math.Abs(improvementList[j].FlatChangePercent)
	})

	fmt.Printf("Total Functions Analyzed: %d\n", len(report.FunctionChanges))
	fmt.Printf("Regressions: %d\n", len(regressionList))
	fmt.Printf("Improvements: %d\n", len(improvementList))
	fmt.Printf("Stable: %d\n", stable)

	if len(regressionList) > 0 {
		fmt.Println("\n⚠️  Top Regressions (worst first):")
		for _, change := range regressionList {
			fmt.Printf("  • %s\n", change.Summary())
		}
	}

	if len(improvementList) > 0 {
		fmt.Println("\n✅ Top Improvements (best first):")
		for _, change := range improvementList {
			fmt.Printf("  • %s\n", change.Summary())
		}
	}
}

func printDetailedReport(report *tracker.ProfileChangeReport) {
	changes := report.FunctionChanges

	// Count each type
	var regressions, improvements, stable int
	for _, change := range changes {
		switch change.ChangeType {
		case "REGRESSION":
			regressions++
		case "IMPROVEMENT":
			improvements++
		default:
			stable++
		}
	}

	// Print header with statistics and sorting info
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Detailed Performance Report                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n📊 Summary: %d total functions | 🔴 %d regressions | 🟢 %d improvements | ⚪ %d stable\n",
		len(changes), regressions, improvements, stable)
	fmt.Println("\n📋 Report Order: Regressions first (worst → best), then Improvements (best → worst), then Stable")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════")

	// Sort by change type first (REGRESSION, IMPROVEMENT, STABLE),
	// then by absolute percentage change (biggest changes first)
	sort.Slice(changes, func(i, j int) bool {
		// Primary sort: by change type priority
		typePriority := map[string]int{
			"REGRESSION":  1,
			"IMPROVEMENT": 2,
			"STABLE":      3,
		}

		if typePriority[changes[i].ChangeType] != typePriority[changes[j].ChangeType] {
			return typePriority[changes[i].ChangeType] < typePriority[changes[j].ChangeType]
		}

		return math.Abs(changes[i].FlatChangePercent) > math.Abs(changes[j].FlatChangePercent)
	})

	for i, change := range changes {
		if i > 0 {
			fmt.Println()
			fmt.Println()
			fmt.Println()
		}
		fmt.Print(change.Report())
	}
}

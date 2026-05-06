package benchmark

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func runBenchAndGetProfiles(runner tooling.Runner, benchArgs *internal.BenchArgs, benchmarkConfigs map[string]internal.FunctionFilter, groupByPackage bool, lenientProfiles bool, skipPNG bool) error {
	slog.Info("Starting benchmark pipeline...")

	var functionFilter internal.FunctionFilter
	globalFilter, hasGlobalFilter := benchmarkConfigs[internal.GlobalSign]
	if hasGlobalFilter {
		functionFilter = globalFilter
	}

	for _, benchmarkName := range benchArgs.Benchmarks {
		slog.Info("Running benchmark", "Benchmark", benchmarkName)
		if err := runBenchmark(runner, benchmarkName, benchArgs.Profiles, benchArgs.Count, benchArgs.Tag); err != nil {
			return fmt.Errorf("failed to run %s: %w", benchmarkName, err)
		}

		slog.Info("Processing profiles", "Benchmark", benchmarkName)
		profilesReady, err := processProfiles(runner, benchmarkName, benchArgs.Profiles, benchArgs.Tag, groupByPackage, lenientProfiles, skipPNG)
		if err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		slog.Info("Analyzing profile functions", "Benchmark", benchmarkName)

		if !hasGlobalFilter {
			functionFilter = benchmarkConfigs[benchmarkName]
		}

		args := &internal.CollectionArgs{
			Tag:             benchArgs.Tag,
			Profiles:        profilesReady,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: functionFilter,
		}

		if collErr := collectProfileFunctions(runner, args); collErr != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, collErr)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info(internal.InfoCollectionSuccess)
	return nil
}

// collectProfileFunctions collects all pprof information for each function, according to configurations.
func collectProfileFunctions(runner tooling.Runner, args *internal.CollectionArgs) error {
	for _, profile := range args.Profiles {
		paths := getProfilePaths(args.Tag, args.BenchmarkName, profile)
		if err := os.MkdirAll(paths.FunctionDirectory, internal.PermDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		listEntries, err := parser.GetFunctionListEntriesV2(paths.ProfileBinaryFile, args.BenchmarkConfig)
		if err != nil {
			return fmt.Errorf("failed to extract function names: %w", err)
		}

		if err = collector.GetFunctionsOutput(runner, listEntries, paths.ProfileBinaryFile, paths.FunctionDirectory); err != nil {
			return fmt.Errorf("getAllFunctionsPprofContents failed: %w", err)
		}
	}

	return nil
}

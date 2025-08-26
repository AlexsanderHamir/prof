package benchmark

import (
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/prof/internal"
)

func runBenchAndGetProfiles(benchArgs *internal.BenchArgs, benchmarkConfigs map[string]internal.FunctionFilter, groupByPackage bool) error {
	slog.Info("Starting benchmark pipeline...")

	var functionFilter internal.FunctionFilter
	globalFilter, hasGlobalFilter := benchmarkConfigs[internal.GlobalSign]
	if hasGlobalFilter {
		functionFilter = globalFilter
	}

	for _, benchmarkName := range benchArgs.Benchmarks {
		slog.Info("Running benchmark", "Benchmark", benchmarkName)
		if err := runBenchmark(benchmarkName, benchArgs.Profiles, benchArgs.Count, benchArgs.Tag); err != nil {
			return fmt.Errorf("failed to run %s: %w", benchmarkName, err)
		}

		slog.Info("Processing profiles", "Benchmark", benchmarkName)
		if err := processProfiles(benchmarkName, benchArgs.Profiles, benchArgs.Tag, groupByPackage); err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		slog.Info("Analyzing profile functions", "Benchmark", benchmarkName)

		if !hasGlobalFilter {
			functionFilter = benchmarkConfigs[benchmarkName]
		}

		args := &internal.CollectionArgs{
			Tag:             benchArgs.Tag,
			Profiles:        benchArgs.Profiles,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: functionFilter,
		}

		if err := collectProfileFunctions(args); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info(internal.InfoCollectionSuccess)
	return nil
}

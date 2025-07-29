package cli

import (
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/prof/engine/benchmark"
	"github.com/AlexsanderHamir/prof/internal/args"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/shared"
)

func printConfiguration(benchArgs *args.BenchArgs, functionFilterPerBench map[string]config.FunctionFilter) {
	slog.Info(
		"Parsed arguments",
		"Benchmarks", benchArgs.Benchmarks,
		"Profiles", benchArgs.Profiles,
		"Tag", benchArgs.Tag,
		"Count", benchArgs.Count,
	)

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

func runBenchAndGetProfiles(benchArgs *args.BenchArgs, benchmarkConfigs map[string]config.FunctionFilter) error {
	slog.Info("Starting benchmark pipeline...")

	var functionFilter config.FunctionFilter
	globalFilter, hasGlobalFilter := benchmarkConfigs[shared.GlobalSign]
	if hasGlobalFilter {
		functionFilter = globalFilter
	}

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

		if !hasGlobalFilter {
			functionFilter = benchmarkConfigs[benchmarkName]
		}

		args := &args.CollectionArgs{
			Tag:             benchArgs.Tag,
			Profiles:        benchArgs.Profiles,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: functionFilter,
		}

		if err := benchmark.CollectProfileFunctions(args); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info(shared.InfoCollectionSuccess)
	return nil
}

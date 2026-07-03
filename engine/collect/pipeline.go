package collect

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

func runBenchAndGetProfiles(runner tooling.Runner, autoArgs *config.AutoArgs, cfg *config.Config, lenientProfiles bool, skipPNG bool) error {
	slog.Info("Starting benchmark pipeline...")

	total := len(autoArgs.Benchmarks)
	for i, benchmarkName := range autoArgs.Benchmarks {
		progress := termui.Progress{
			Label:  benchmarkName,
			Index:  i + 1,
			Total:  total,
			Detail: fmt.Sprintf("count=%d", autoArgs.Count),
		}
		slog.Info("Running benchmark", "Benchmark", benchmarkName)
		if err := runBenchmark(runner, benchmarkName, autoArgs.Profiles, autoArgs.Count, autoArgs.Tag, progress); err != nil {
			return fmt.Errorf("failed to run %s: %w", benchmarkName, err)
		}

		filter := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto(benchmarkName))

		slog.Info("Processing profiles", "Benchmark", benchmarkName)
		profilesReady, err := processProfiles(runner, benchmarkName, autoArgs.Profiles, autoArgs.Tag, lenientProfiles, skipPNG)
		if err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		slog.Info("Analyzing profile functions", "Benchmark", benchmarkName)

		args := &config.CollectionArgs{
			Tag:             autoArgs.Tag,
			Profiles:        profilesReady,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: filter,
		}

		if collErr := collectProfileFunctions(runner, args); collErr != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, collErr)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info(workspace.InfoCollectionSuccess)
	return nil
}

func collectProfileFunctions(runner tooling.Runner, args *config.CollectionArgs) error {
	layout, err := workspace.TagLayoutFromCWD(args.Tag)
	if err != nil {
		return err
	}

	for _, profile := range args.Profiles {
		fnDir := layout.FunctionsDir(profile, args.BenchmarkName)
		if mkdirErr := os.MkdirAll(fnDir, workspace.PermDir); mkdirErr != nil {
			return fmt.Errorf("failed to create output directory: %w", mkdirErr)
		}

		binPath := layout.Bin(args.BenchmarkName, profile)
		listEntries, listErr := parser.GetFunctionListEntriesV2(binPath, args.BenchmarkConfig)
		if listErr != nil {
			return fmt.Errorf("failed to extract function names: %w", listErr)
		}

		if fnErr := getFunctionsOutput(runner, listEntries, binPath, fnDir); fnErr != nil {
			return fmt.Errorf("getAllFunctionsPprofContents failed: %w", fnErr)
		}
	}

	return nil
}

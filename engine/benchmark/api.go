package benchmark

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func RunBenchmarks(benchmarks, profiles []string, tag string, count int) error {
	if len(benchmarks) == 0 {
		return errors.New("benchmarks flag is empty")
	}

	if len(profiles) == 0 {
		return errors.New("profiles flag is empty")
	}

	cfg, err := internal.LoadFromFile(internal.ConfigFilename)
	if err != nil {
		slog.Info("No config file found at repository root; proceeding without function filters.", "expected", internal.ConfigFilename)
		slog.Info("You can generate one with 'prof setup'. It will be placed at the root next to go.mod.")
		cfg = &internal.Config{}
	}

	if err = SetupDirectories(tag, benchmarks, profiles); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	benchArgs := &internal.BenchArgs{
		Benchmarks: benchmarks,
		Profiles:   profiles,
		Count:      count,
		Tag:        tag,
	}

	internal.PrintConfiguration(benchArgs, cfg.FunctionFilter)

	if err = RunBenchAndGetProfiles(benchArgs, cfg.FunctionFilter); err != nil {
		return err
	}

	return nil
}

func RunBenchAndGetProfiles(benchArgs *internal.BenchArgs, benchmarkConfigs map[string]internal.FunctionFilter) error {
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
		if err := ProcessProfiles(benchmarkName, benchArgs.Profiles, benchArgs.Tag); err != nil {
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

		if err := CollectProfileFunctions(args); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info(internal.InfoCollectionSuccess)
	return nil
}

// SetupDirectories creates the structure of the library's output.
func SetupDirectories(tag string, benchmarks, profiles []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	tagDir := filepath.Join(currentDir, internal.MainDirOutput, tag)
	err = internal.CleanOrCreateDir(tagDir)
	if err != nil {
		return fmt.Errorf("CleanOrCreateDir failed: %w", err)
	}

	if err = createBenchDirectories(tagDir, benchmarks); err != nil {
		return err
	}

	return createProfileFunctionDirectories(tagDir, profiles, benchmarks)
}

// CollectProfileFunctions collects all pprof information for each function, according to configurations.
func CollectProfileFunctions(args *internal.CollectionArgs) error {
	for _, profile := range args.Profiles {
		paths := getProfilePaths(args.Tag, args.BenchmarkName, profile)
		if err := os.MkdirAll(paths.FunctionDirectory, internal.PermDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		functions, err := parser.GetAllFunctionNames(paths.ProfileTextFile, args.BenchmarkConfig)
		if err != nil {
			return fmt.Errorf("failed to extract function names: %w", err)
		}

		if err = collector.GetFunctionsOutput(functions, paths.ProfileBinaryFile, paths.FunctionDirectory); err != nil {
			return fmt.Errorf("getAllFunctionsPprofContents failed: %w", err)
		}
	}

	return nil
}

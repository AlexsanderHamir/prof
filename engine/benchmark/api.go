package benchmark

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/internal/args"
	"github.com/AlexsanderHamir/prof/internal/shared"
	"github.com/AlexsanderHamir/prof/parser"
)

// SetupDirectories creates the structure of the library's output.
func SetupDirectories(tag string, benchmarks, profiles []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	tagDir := filepath.Join(currentDir, shared.MainDirOutput, tag)
	err = shared.CleanOrCreateDir(tagDir)
	if err != nil {
		return fmt.Errorf("CleanOrCreateDir failed: %w", err)
	}

	if err = createBenchDirectories(tagDir, benchmarks); err != nil {
		return err
	}

	return createProfileFunctionDirectories(tagDir, profiles, benchmarks)
}

// RunBenchmark runs a specific benchmark and collects all of its information.
func RunBenchmark(benchmarkName string, profiles []string, count int, tag string) error {
	cmd, err := buildBenchmarkCommand(benchmarkName, profiles, count)
	if err != nil {
		return err
	}

	textDir, binDir := getOutputDirectories(benchmarkName, tag)

	outputFile := filepath.Join(textDir, fmt.Sprintf("%s.%s", benchmarkName, shared.TextExtension))
	if err = runBenchmarkCommand(cmd, outputFile); err != nil {
		return err
	}

	if err = moveProfileFiles(benchmarkName, profiles, binDir); err != nil {
		return err
	}

	return moveTestFiles(benchmarkName, binDir)
}

// ProcessProfiles collects all pprof info for a specific benchmark and its specified profiles.
func ProcessProfiles(benchmarkName string, profiles []string, tag string) error {
	tagDir := filepath.Join(shared.MainDirOutput, tag)
	binDir := filepath.Join(tagDir, shared.ProfileBinDir, benchmarkName)
	textDir := filepath.Join(tagDir, shared.ProfileTextDir, benchmarkName)

	for _, profile := range profiles {
		profileFile := filepath.Join(binDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, binExtension))
		if _, err := os.Stat(profileFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.Warn("Profile file not found", "file", profileFile)
				continue
			}
			return fmt.Errorf("failed to stat profile file %s: %w", profileFile, err)
		}

		outputFile := filepath.Join(textDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, shared.TextExtension))
		profileFunctionsDir := filepath.Join(tagDir, profile+shared.FunctionsDirSuffix, benchmarkName)

		if err := collector.GetProfileTextOutput(profileFile, outputFile); err != nil {
			return fmt.Errorf("failed to generate text profile for %s: %w", profile, err)
		}

		pngDesiredFilePath := filepath.Join(profileFunctionsDir, fmt.Sprintf("%s_%s.png", benchmarkName, profile))
		if err := collector.GetPNGOutput(profileFile, pngDesiredFilePath); err != nil {
			return fmt.Errorf("failed to generate PNG visualization for %s: %w", profile, err)
		}

		slog.Info("Processed profile", "profile", profile, "benchmark", benchmarkName)
	}

	return nil
}

// CollectProfileFunctions collects all pprof information for each function, according to configurations.
func CollectProfileFunctions(args *args.CollectionArgs) error {
	for _, profile := range args.Profiles {
		paths := getProfilePaths(args.Tag, args.BenchmarkName, profile)
		if err := os.MkdirAll(paths.FunctionDirectory, shared.PermDir); err != nil {
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

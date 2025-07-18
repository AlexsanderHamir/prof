package benchmark

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
)

// SetupDirectories creates the structure of the library's output.
func SetupDirectories(tag string, benchmarks, profiles []string) error {
	if err := createBenchDirectories(tag, benchmarks); err != nil {
		return err
	}
	return createProfileFunctionDirectories(tag, profiles, benchmarks)
}

// RunBenchmark runs a specific benchmark and collects all of its information.
func RunBenchmark(benchmarkName string, profiles []string, count int, tag string) error {
	cmd := buildBenchmarkCommand(benchmarkName, profiles, count)
	textDir, binDir, err := getOrCreateOutputDirectories(benchmarkName, tag)
	if err != nil {
		return fmt.Errorf("couldn't get output directories: %w", err)
	}

	outputFile := filepath.Join(textDir, fmt.Sprintf("%s.%s", benchmarkName, textExtension))
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
		if profile == shared.TRACE {
			continue
		}

		profileFile := filepath.Join(binDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, binExtension))
		if _, err := os.Stat(profileFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.Warn("Profile file not found", "file", profileFile)
				continue
			}
			return fmt.Errorf("failed to stat profile file %s: %w", profileFile, err)
		}

		outputFile := filepath.Join(textDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, textExtension))
		profileFunctionsDir := filepath.Join(tagDir, profile+shared.FunctionsDirSuffix, benchmarkName)

		if err := generateTextProfile(profileFile, outputFile); err != nil {
			return fmt.Errorf("failed to generate text profile for %s: %w", profile, err)
		}

		pngFile := filepath.Join(profileFunctionsDir, fmt.Sprintf("%s_%s.png", benchmarkName, profile))
		if err := generatePNGVisualization(profileFile, pngFile); err != nil {
			return fmt.Errorf("failed to generate PNG visualization for %s: %w", profile, err)
		}

		slog.Info("Processed profile", "profile", profile, "benchmark", benchmarkName)
	}

	return nil
}

// CollectProfileFunctions collects all pprof information for each function, according to configurations.
func CollectProfileFunctions(args *args.CollectionArgs) error {
	for _, profile := range args.Profiles {
		if profile == shared.TRACE {
			continue
		}

		paths := getProfilePaths(args.Tag, args.BenchmarkName, profile)
		if err := os.MkdirAll(paths.FunctionDirectory, shared.PermDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		filter := parser.ProfileFilter{
			FunctionPrefixes: args.BenchmarkConfig.IncludePrefixes,
			IgnoreFunctions:  args.BenchmarkConfig.IgnoreFunctions,
		}

		functions, err := parser.GetAllFunctionNames(paths.ProfileTextFile, filter)
		if err != nil {
			return fmt.Errorf("failed to extract function names: %w", err)
		}

		if err = saveAllFunctionsPprofContents(functions, paths); err != nil {
			return fmt.Errorf("getAllFunctionsPprofContents failed: %w", err)
		}
	}

	return nil
}

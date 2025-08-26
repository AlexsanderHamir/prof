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

// ProfilePaths holds paths for profile text, binary, and output directories.
type ProfilePaths struct {
	// Desired file path for specified profile
	ProfileTextFile string

	// Desired bin path for specified profile
	ProfileBinaryFile string

	// Desired benchmark directory for function data collection
	FunctionDirectory string
}

// getProfilePaths constructs file paths for benchmark profile data organized by tag and benchmark.
//
// Returns paths for:
//   - ProfileTextFile: bench/{tag}/text/{benchmarkName}/{benchmarkName}_{profile}.txt
//   - ProfileBinaryFile: bench/{tag}/bin/{benchmarkName}/{benchmarkName}_{profile}.out
//   - FunctionDirectory: bench/{tag}/{profile}_functions/{benchmarkName}/
//
// Example with tag="v1.0", benchmarkName="BenchmarkPool", profile="cpu":
//   - bench/v1.0/text/BenchmarkPool/BenchmarkPool_cpu.txt
//   - bench/v1.0/bin/BenchmarkPool/BenchmarkPool_cpu.out
//   - bench/v1.0/cpu_functions/BenchmarkPool/function1.txt
func getProfilePaths(tag, benchmarkName, profile string) ProfilePaths {
	tagDir := filepath.Join(internal.MainDirOutput, tag)
	profileTextFile := fmt.Sprintf("%s_%s.%s", benchmarkName, profile, internal.TextExtension)
	profileBinFile := fmt.Sprintf("%s_%s.%s", benchmarkName, profile, binExtension)

	return ProfilePaths{
		ProfileTextFile:   filepath.Join(tagDir, internal.ProfileTextDir, benchmarkName, profileTextFile),
		ProfileBinaryFile: filepath.Join(tagDir, internal.ProfileBinDir, benchmarkName, profileBinFile),
		FunctionDirectory: filepath.Join(tagDir, profile+internal.FunctionsDirSuffix, benchmarkName),
	}
}

// processProfiles collects all pprof info for a specific benchmark and its specified profiles.
func processProfiles(benchmarkName string, profiles []string, tag string, groupByPackage bool) error {
	tagDir := filepath.Join(internal.MainDirOutput, tag)
	binDir := filepath.Join(tagDir, internal.ProfileBinDir, benchmarkName)
	textDir := filepath.Join(tagDir, internal.ProfileTextDir, benchmarkName)

	for _, profile := range profiles {
		profileFile := filepath.Join(binDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, binExtension))
		if _, err := os.Stat(profileFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.Warn("Profile file not found", "file", profileFile)
				continue
			}
			return fmt.Errorf("failed to stat profile file %s: %w", profileFile, err)
		}

		outputFile := filepath.Join(textDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, internal.TextExtension))
		profileFunctionsDir := filepath.Join(tagDir, profile+internal.FunctionsDirSuffix, benchmarkName)

		if err := collector.GetProfileTextOutput(profileFile, outputFile); err != nil {
			return fmt.Errorf("failed to generate text profile for %s: %w", profile, err)
		}

		// Generate grouped profile data if requested
		if groupByPackage {
			groupedOutputFile := filepath.Join(textDir, fmt.Sprintf("%s_%s_grouped.%s", benchmarkName, profile, internal.TextExtension))
			if err := generateGroupedProfileData(profileFile, groupedOutputFile, internal.FunctionFilter{}); err != nil {
				return fmt.Errorf("failed to generate grouped profile for %s: %w", profile, err)
			}
		}

		pngDesiredFilePath := filepath.Join(profileFunctionsDir, fmt.Sprintf("%s_%s.png", benchmarkName, profile))
		if err := collector.GetPNGOutput(profileFile, pngDesiredFilePath); err != nil {
			return fmt.Errorf("failed to generate PNG visualization for %s: %w", profile, err)
		}

		slog.Info("Processed profile", "profile", profile, "benchmark", benchmarkName)
	}

	return nil
}

// generateGroupedProfileData generates profile data organized by package/module using the new parser function
func generateGroupedProfileData(binaryFile, outputFile string, functionFilter internal.FunctionFilter) error {
	// Import the parser package to use OrganizeProfileByPackageV2
	groupedData, err := parser.OrganizeProfileByPackageV2(binaryFile, functionFilter)
	if err != nil {
		return fmt.Errorf("failed to organize profile by package: %w", err)
	}

	// Write the grouped data to the output file
	return os.WriteFile(outputFile, []byte(groupedData), internal.PermFile)
}

// CollectProfileFunctions collects all pprof information for each function, according to configurations.
func collectProfileFunctions(args *internal.CollectionArgs) error {
	for _, profile := range args.Profiles {
		paths := getProfilePaths(args.Tag, args.BenchmarkName, profile)
		if err := os.MkdirAll(paths.FunctionDirectory, internal.PermDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		functions, err := parser.GetAllFunctionNamesV2(paths.ProfileBinaryFile, args.BenchmarkConfig)
		if err != nil {
			return fmt.Errorf("failed to extract function names: %w", err)
		}

		if err = collector.GetFunctionsOutput(functions, paths.ProfileBinaryFile, paths.FunctionDirectory); err != nil {
			return fmt.Errorf("getAllFunctionsPprofContents failed: %w", err)
		}
	}

	return nil
}

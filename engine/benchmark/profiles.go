package benchmark

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal"
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
	profileBinFile := fmt.Sprintf("%s_%s.%s", benchmarkName, profile, internal.ProfileArtifactExtension)

	return ProfilePaths{
		ProfileTextFile:   filepath.Join(tagDir, internal.ProfileTextDir, benchmarkName, profileTextFile),
		ProfileBinaryFile: filepath.Join(tagDir, internal.ProfileBinDir, benchmarkName, profileBinFile),
		FunctionDirectory: filepath.Join(tagDir, profile+internal.FunctionsDirSuffix, benchmarkName),
	}
}

// processProfiles collects all pprof info for a specific benchmark and its specified profiles.
// It returns profile kinds successfully processed—when lenientProfiles is true, missing binaries are omitted
// from this slice so downstream collection skips them consistently.
//
// Processing runs in order: text listing, optional package-grouped text, then PNG (optional when skipPNG is true).
func processProfiles(runner tooling.Runner, benchmarkName string, profiles []string, tag string, groupByPackage bool, lenientProfiles bool, skipPNG bool) ([]string, error) { //nolint:gocognit // sequential profile stages
	tagDir := filepath.Join(internal.MainDirOutput, tag)
	binDir := filepath.Join(tagDir, internal.ProfileBinDir, benchmarkName)
	textDir := filepath.Join(tagDir, internal.ProfileTextDir, benchmarkName)

	var processed []string

	for _, profile := range profiles {
		profileFile := filepath.Join(binDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, internal.ProfileArtifactExtension))
		if _, err := os.Stat(profileFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if lenientProfiles {
					slog.Warn("Profile file not found — skipping", "file", profileFile)
					continue
				}
				return nil, fmt.Errorf("missing profile binary for benchmark %s profile %s: %w", benchmarkName, profile, err)
			}
			return nil, fmt.Errorf("failed to stat profile file %s: %w", profileFile, err)
		}

		outputFile := filepath.Join(textDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, internal.TextExtension))
		profileFunctionsDir := filepath.Join(tagDir, profile+internal.FunctionsDirSuffix, benchmarkName)

		if err := collector.GetProfileTextOutput(runner, profileFile, outputFile); err != nil {
			return nil, fmt.Errorf("failed to generate text profile for %s: %w", profile, err)
		}

		if groupByPackage {
			groupedOutputFile := filepath.Join(textDir, fmt.Sprintf("%s_%s_grouped.%s", benchmarkName, profile, internal.TextExtension))
			if err := collector.WriteGroupedPackageProfile(profileFile, groupedOutputFile, internal.FunctionFilter{}); err != nil {
				return nil, fmt.Errorf("failed to generate grouped profile for %s: %w", profile, err)
			}
		}

		if err := os.MkdirAll(profileFunctionsDir, internal.PermDir); err != nil {
			return nil, fmt.Errorf("failed to create profile functions directory: %w", err)
		}

		pngDesiredFilePath := filepath.Join(profileFunctionsDir, fmt.Sprintf("%s_%s.png", benchmarkName, profile))
		if err := collector.GetPNGOutput(runner, profileFile, pngDesiredFilePath); err != nil {
			if skipPNG {
				slog.Warn("PNG visualization skipped", "profile", profile, "benchmark", benchmarkName, "err", err)
			} else {
				return nil, fmt.Errorf("failed to generate PNG for profile %s (install graphviz or use --skip-png): %w", profile, err)
			}
		}

		slog.Info("Processed profile", "profile", profile, "benchmark", benchmarkName)
		processed = append(processed, profile)
	}

	if len(processed) == 0 && len(profiles) > 0 {
		return nil, fmt.Errorf("no profile binaries processed for benchmark %s", benchmarkName)
	}

	return processed, nil
}

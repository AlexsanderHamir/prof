package analyzer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/shared"
)

// ValidateBenchmarkDirectories checks if the benchmark directories exist for a given tag and returns the benchmark names.
// If specificBenchmarks is nil or empty, returns all valid benchmark directories. Otherwise, validates and returns only those that exist.
func ValidateBenchmarkDirectories(tag string, specificBenchmarks []string) ([]string, error) {
	baseDir := filepath.Join(shared.Main_dir_output, tag)

	if _, err := os.Stat(baseDir); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no benchmark data found for tag '%s'", tag)
	}

	textDir := filepath.Join(baseDir, shared.Profile_text_files_directory)
	if _, err := os.Stat(textDir); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no text profiles found in %s", textDir)
	}

	entries, err := os.ReadDir(textDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read text directory: %w", err)
	}

	var allBenchmarkNames []string
	benchmarkSet := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			allBenchmarkNames = append(allBenchmarkNames, name)
			benchmarkSet[name] = struct{}{}
		}
	}

	if len(allBenchmarkNames) == 0 {
		return nil, fmt.Errorf("no benchmark directories found in %s", textDir)
	}

	// If no specific benchmarks requested, return all
	if len(specificBenchmarks) == 0 {
		return allBenchmarkNames, nil
	}

	// Validate specific benchmarks
	var validBenchmarks []string
	var missing []string
	for _, name := range specificBenchmarks {
		if _, ok := benchmarkSet[name]; ok {
			validBenchmarks = append(validBenchmarks, name)
		} else {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("the following benchmarks were not found in %s: %v", textDir, missing)
	}

	return validBenchmarks, nil
}

// AnalyzeAllProfiles runs analysis for all benchmarks and profile types for a given tag.
func AnalyzeAllProfiles(tag string, benchmarkNames, profileTypes []string, cfg *config.Config, isFlagging bool) error {
	log.Printf("\nStarting comprehensive analysis for tag: %s\n", tag)
	log.Printf("Benchmarks: %v\n", benchmarkNames)
	log.Printf("Profile types: %v\n", profileTypes)
	log.Printf("================================================================================\n")

	for _, benchmarkName := range benchmarkNames {
		for _, profileType := range profileTypes {
			if profileType == shared.TRACE {
				continue
			}

			log.Printf("\nAnalyzing %s (%s)...\n", benchmarkName, profileType)
			if err := sendToModel(tag, benchmarkName, profileType, cfg, isFlagging); err != nil {
				return fmt.Errorf("failed to analyze %s (%s): %w", benchmarkName, profileType, err)
			}
		}
	}

	return nil
}

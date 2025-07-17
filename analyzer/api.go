package analyzer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/config"
)

const (
	permDir  = 0o755
	permFile = 0o644
)

// ValidateBenchmarkDirectories checks if the benchmark directories exist for a given tag and returns the benchmark names.
func ValidateBenchmarkDirectories(tag string) ([]string, error) {
	baseDir := filepath.Join("bench", tag)

	if _, err := os.Stat(baseDir); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no benchmark data found for tag '%s'", tag)
	}

	textDir := filepath.Join(baseDir, "text")
	if _, err := os.Stat(textDir); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no text profiles found in %s", textDir)
	}

	entries, err := os.ReadDir(textDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read text directory: %w", err)
	}

	var benchmarkNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			benchmarkNames = append(benchmarkNames, entry.Name())
		}
	}

	if len(benchmarkNames) == 0 {
		return nil, fmt.Errorf("no benchmark directories found in %s", textDir)
	}

	return benchmarkNames, nil
}

// AnalyzeAllProfiles runs analysis for all benchmarks and profile types for a given tag.
func AnalyzeAllProfiles(tag string, benchmarkNames, profileTypes []string, cfg *config.Config, isFlag bool) error {
	log.Printf("\nStarting comprehensive analysis for tag: %s\n", tag)
	log.Printf("Benchmarks: %v\n", benchmarkNames)
	log.Printf("Profile types: %v\n", profileTypes)
	log.Printf("================================================================================\n")

	for _, benchmarkName := range benchmarkNames {
		for _, profileType := range profileTypes {
			if profileType == "trace" {
				continue
			}

			log.Printf("\nAnalyzing %s (%s)...\n", benchmarkName, profileType)
			if err := sendToModel(tag, benchmarkName, profileType, cfg, isFlag); err != nil {
				return fmt.Errorf("failed to analyze %s (%s): %w", benchmarkName, profileType, err)
			}
		}
	}

	return nil
}

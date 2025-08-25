package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/internal"
)

// discoverAvailableTags scans the bench directory for existing tags
func discoverAvailableTags() ([]string, error) {
	root, err := internal.FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root: %w", err)
	}

	benchDir := filepath.Join(root, internal.MainDirOutput)
	entries, err := os.ReadDir(benchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read bench directory: %w", err)
	}

	var tags []string
	for _, entry := range entries {
		if entry.IsDir() {
			tags = append(tags, entry.Name())
		}
	}

	return tags, nil
}

// discoverAvailableBenchmarks scans a specific tag directory for available benchmarks
func discoverAvailableBenchmarks(tag string) ([]string, error) {
	root, err := internal.FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root: %w", err)
	}

	benchDir := filepath.Join(root, internal.MainDirOutput, tag, internal.ProfileTextDir)
	entries, err := os.ReadDir(benchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read benchmark directory for tag %s: %w", tag, err)
	}

	var availableBenchmarks []string
	for _, entry := range entries {
		if entry.IsDir() {
			availableBenchmarks = append(availableBenchmarks, entry.Name())
		}
	}

	return availableBenchmarks, nil
}

// discoverAvailableProfiles scans a specific tag and benchmark for available profile types
func discoverAvailableProfiles(tag, benchmarkName string) ([]string, error) {
	root, err := internal.FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root: %w", err)
	}

	benchDir := filepath.Join(root, internal.MainDirOutput, tag, internal.ProfileTextDir, benchmarkName)
	entries, err := os.ReadDir(benchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read profile directory for tag %s, benchmark %s: %w", tag, benchmarkName, err)
	}

	var availableProfiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			// Extract profile type from filename like "BenchmarkName_cpu.txt"
			name := entry.Name()
			if strings.HasPrefix(name, benchmarkName+"_") {
				profileTypeName := strings.TrimSuffix(strings.TrimPrefix(name, benchmarkName+"_"), ".txt")
				if profileTypeName == "cpu" || profileTypeName == "memory" || profileTypeName == "mutex" || profileTypeName == "block" {
					availableProfiles = append(availableProfiles, profileTypeName)
				}
			}
		}
	}

	return availableProfiles, nil
}

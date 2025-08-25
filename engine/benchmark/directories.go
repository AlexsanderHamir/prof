package benchmark

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
)

// createBenchDirectories creates the main structure of the library's output.
func createBenchDirectories(tagDir string, benchmarks []string) error {
	binDir := filepath.Join(tagDir, internal.ProfileBinDir)
	textDir := filepath.Join(tagDir, internal.ProfileTextDir)
	descFile := filepath.Join(tagDir, descriptionFileName)

	// Create main directories
	if err := os.Mkdir(binDir, internal.PermDir); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	if err := os.Mkdir(textDir, internal.PermDir); err != nil {
		return fmt.Errorf("failed to create text directory: %w", err)
	}

	// Create benchmark subdirectories
	for _, benchmark := range benchmarks {
		if err := os.Mkdir(filepath.Join(binDir, benchmark), internal.PermDir); err != nil {
			return fmt.Errorf("failed to create bin subdirectory for %s: %w", benchmark, err)
		}
		if err := os.Mkdir(filepath.Join(textDir, benchmark), internal.PermDir); err != nil {
			return fmt.Errorf("failed to create text subdirectory for %s: %w", benchmark, err)
		}
	}

	// Create description file
	if err := os.WriteFile(descFile, []byte(descritpionFileMessage), internal.PermFile); err != nil {
		return fmt.Errorf("failed to create description file: %w", err)
	}

	slog.Info("Created directory structure", "dir", tagDir)
	return nil
}

// createProfileFunctionDirectories creates the structure for the code line level data collection.
func createProfileFunctionDirectories(tagDir string, profiles, benchmarks []string) error {
	for _, profileName := range profiles {
		profileDirPath := filepath.Join(tagDir, profileName+internal.FunctionsDirSuffix)
		if err := os.Mkdir(profileDirPath, internal.PermDir); err != nil {
			return fmt.Errorf("failed to create profile directory %s: %w", profileDirPath, err)
		}

		for _, benchmark := range benchmarks {
			benchmarkDirPath := filepath.Join(profileDirPath, benchmark)
			if err := os.Mkdir(benchmarkDirPath, internal.PermDir); err != nil {
				return fmt.Errorf("failed to create benchmark directory %s: %w", benchmarkDirPath, err)
			}
		}
	}

	slog.Info("Created profile function directories")
	return nil
}

// SetupDirectories creates the structure of the library's output.
func setupDirectories(tag string, benchmarks, profiles []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	tagDir := filepath.Join(currentDir, internal.MainDirOutput, tag)
	err = internal.CleanOrCreateTag(tagDir)
	if err != nil {
		return fmt.Errorf("CleanOrCreateTag failed: %w", err)
	}

	if err = createBenchDirectories(tagDir, benchmarks); err != nil {
		return err
	}

	return createProfileFunctionDirectories(tagDir, profiles, benchmarks)
}

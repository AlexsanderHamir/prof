package collect

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func createBenchDirectories(tagDir string, benchmarks []string, quiet bool) error {
	binDir := filepath.Join(tagDir, workspace.ProfileBinDir)
	textDir := filepath.Join(tagDir, workspace.ProfileTextDir)
	descFile := filepath.Join(tagDir, workspace.BenchDescriptionFileName)

	if err := os.Mkdir(binDir, workspace.PermDir); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	if err := os.Mkdir(textDir, workspace.PermDir); err != nil {
		return fmt.Errorf("failed to create text directory: %w", err)
	}

	for _, b := range benchmarks {
		if err := os.Mkdir(filepath.Join(binDir, b), workspace.PermDir); err != nil {
			return fmt.Errorf("failed to create bin subdirectory for %s: %w", b, err)
		}
		if err := os.Mkdir(filepath.Join(textDir, b), workspace.PermDir); err != nil {
			return fmt.Errorf("failed to create text subdirectory for %s: %w", b, err)
		}
	}

	if err := os.WriteFile(descFile, []byte(workspace.BenchDescriptionPlaceholder), workspace.PermFile); err != nil {
		return fmt.Errorf("failed to create description file: %w", err)
	}

	if !quiet {
		slog.Info("Created directory structure", "dir", tagDir)
	}
	return nil
}

func createProfileFunctionDirectories(tagDir string, profiles, benchmarks []string, quiet bool) error {
	for _, profileName := range profiles {
		profileDirPath := filepath.Join(tagDir, profileName+workspace.FunctionsDirSuffix)
		if err := os.Mkdir(profileDirPath, workspace.PermDir); err != nil {
			return fmt.Errorf("failed to create profile directory %s: %w", profileDirPath, err)
		}
		for _, b := range benchmarks {
			benchmarkDirPath := filepath.Join(profileDirPath, b)
			if err := os.Mkdir(benchmarkDirPath, workspace.PermDir); err != nil {
				return fmt.Errorf("failed to create benchmark directory %s: %w", benchmarkDirPath, err)
			}
		}
	}
	if !quiet {
		slog.Info("Created profile function directories")
	}
	return nil
}

func setupDirectories(tag string, benchmarks, profiles []string, quiet bool) error {
	tagDir, err := workspace.TagDirFromCWD(tag)
	if err != nil {
		return err
	}
	if err = workspace.CleanOrCreateTag(tagDir); err != nil {
		return fmt.Errorf("CleanOrCreateTag failed: %w", err)
	}
	if err = createBenchDirectories(tagDir, benchmarks, quiet); err != nil {
		return err
	}
	return createProfileFunctionDirectories(tagDir, profiles, benchmarks, quiet)
}

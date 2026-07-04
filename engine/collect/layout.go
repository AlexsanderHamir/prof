package collect

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func createBenchDirectories(tagDir string, benchmarks []string, quiet bool) error {
	profilesDir := filepath.Join(tagDir, workspace.ProfilesDir)
	measurementsDir := filepath.Join(tagDir, workspace.MeasurementsDir)
	hotspotsDir := filepath.Join(tagDir, workspace.HotspotsDir)
	notesFile := filepath.Join(tagDir, workspace.TagNotesFileName)

	for _, dir := range []string{profilesDir, measurementsDir, hotspotsDir} {
		if err := os.Mkdir(dir, workspace.PermDir); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", filepath.Base(dir), err)
		}
	}

	for _, b := range benchmarks {
		for _, dir := range []string{profilesDir, measurementsDir, hotspotsDir} {
			if err := os.Mkdir(filepath.Join(dir, b), workspace.PermDir); err != nil {
				return fmt.Errorf("failed to create %s subdirectory for %s: %w", filepath.Base(dir), b, err)
			}
		}
	}

	if err := os.WriteFile(notesFile, []byte(workspace.TagNotesPlaceholder), workspace.PermFile); err != nil {
		return fmt.Errorf("failed to create notes file: %w", err)
	}

	if !quiet {
		slog.Info("Created directory structure", "dir", tagDir)
	}
	return nil
}

func createSourceLinesDirectories(tagDir string, profiles, benchmarks []string, quiet bool) error {
	sourceLinesRoot := filepath.Join(tagDir, workspace.SourceLinesDir)
	if err := os.Mkdir(sourceLinesRoot, workspace.PermDir); err != nil {
		return fmt.Errorf("failed to create source_lines directory: %w", err)
	}
	for _, profileName := range profiles {
		profileRoot := filepath.Join(sourceLinesRoot, profileName)
		if err := os.Mkdir(profileRoot, workspace.PermDir); err != nil {
			return fmt.Errorf("failed to create source_lines/%s directory: %w", profileName, err)
		}
		for _, b := range benchmarks {
			benchmarkDirPath := filepath.Join(profileRoot, b)
			if err := os.Mkdir(benchmarkDirPath, workspace.PermDir); err != nil {
				return fmt.Errorf("failed to create benchmark directory %s: %w", benchmarkDirPath, err)
			}
		}
	}
	if !quiet {
		slog.Info("Created source_lines directories")
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
	return createSourceLinesDirectories(tagDir, profiles, benchmarks, quiet)
}

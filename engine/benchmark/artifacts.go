package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexsanderHamir/prof/internal"
)

func getExpectedProfileFileName(profile string) (string, bool) {
	expectedFileName, exists := ExpectedFiles[profile]
	if !exists {
		return "", false
	}
	return expectedFileName, true
}

func findMostRecentFile(rootDir, fileName string) (string, error) {
	var latestPath string
	var latestMod time.Time
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != fileName {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return statErr
		}
		if info.ModTime().After(latestMod) {
			latestMod = info.ModTime()
			latestPath = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return latestPath, nil
}

// isGoTestBinary reports whether name is a Go test executable left in the package
// directory after "go test". On Windows the binary is named *.test.exe; elsewhere *.test.
func isGoTestBinary(name string) bool {
	if strings.HasSuffix(name, ".test") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(name), ".test.exe")
}

func moveProfileFiles(benchmarkName string, profiles []string, rootDir string, binDir string) error {
	for _, profile := range profiles {
		profileFile, ok := getExpectedProfileFileName(profile)
		if !ok {
			continue
		}
		latestPath, err := findMostRecentFile(rootDir, profileFile)
		if err != nil {
			return fmt.Errorf("failed to search for profile files: %w", err)
		}
		if latestPath == "" {
			continue
		}
		destPath := filepath.Join(binDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, internal.ProfileArtifactExtension))
		if err = os.Rename(latestPath, destPath); err != nil {
			return fmt.Errorf("failed to move profile file %s: %w", latestPath, err)
		}
	}
	return nil
}

func moveTestFiles(benchmarkName, rootDir, binDir string) error {
	var testFiles []string
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if isGoTestBinary(d.Name()) {
			testFiles = append(testFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("WalkDir Failed: %w", err)
	}
	for _, file := range testFiles {
		newPath := filepath.Join(binDir, fmt.Sprintf("%s_%s", benchmarkName, filepath.Base(file)))
		if err = os.Rename(file, newPath); err != nil {
			return fmt.Errorf("failed to move test file %s: %w", file, err)
		}
	}
	return nil
}

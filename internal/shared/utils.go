package shared

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CLI Commands
const (
	AUTOCMD        = "auto"
	MANUALCMD      = "manual"
	TrackAutoCMD   = AUTOCMD
	TrackManualCMD = MANUALCMD
)

const (
	InfoCollectionSuccess = "All benchmarks and profile processing completed successfully!"
	IMPROVEMENT           = "IMPROVEMENT"
	REGRESSION            = "REGRESSION"
	STABLE                = "STABLE"
)

const (
	MainDirOutput      = "bench"
	ProfileTextDir     = "text"
	ProfileBinDir      = "bin"
	PermDir            = 0o755
	PermFile           = 0o644
	FunctionsDirSuffix = "_functions"
	TextExtension      = "txt"
	ConfigFilename     = "config_template.json"
	GlobalSign         = "*"
)

func GetScanner(filePath string) (*bufio.Scanner, *os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read profile file %s: %w", filePath, err)
	}

	scanner := bufio.NewScanner(file)

	return scanner, file, nil
}

// CleanOrCreateDir cleans a directory if it exists, or creates one if it.
func CleanOrCreateDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, PermDir); err != nil {
				return fmt.Errorf("failed to create %s directory: %w", dir, err)
			}
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err = os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// FindGoModuleRoot searches upwards from the current working directory for a directory
// containing a go.mod file and returns its absolute path. If none is found, an error is returned.
func FindGoModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found from current directory upwards")
		}
		dir = parent
	}
}

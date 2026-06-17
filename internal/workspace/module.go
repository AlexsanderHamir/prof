package workspace

import (
	"errors"
	"os"
	"path/filepath"
)

// FindModuleRoot searches upward from cwd for a directory containing go.mod.
func FindModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Abs(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found from current directory upwards")
		}
		dir = parent
	}
}

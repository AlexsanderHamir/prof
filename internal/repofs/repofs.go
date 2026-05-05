// Package repofs holds repository filesystem helpers: locating the Go module root and tag directories.
package repofs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CleanOrCreateTag cleans the tag directory if it exists, or creates one.
func CleanOrCreateTag(dir string, permDir os.FileMode) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, permDir); err != nil {
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
// containing go.mod and returns its absolute path. If none is found, an error is returned.
func FindGoModuleRoot() (string, error) {
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

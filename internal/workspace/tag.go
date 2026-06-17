package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// CleanOrCreateTag cleans the tag directory if it exists, or creates one.
func CleanOrCreateTag(dir string) error {
	return cleanOrCreateTag(dir, PermDir)
}

func cleanOrCreateTag(dir string, permDir os.FileMode) error {
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

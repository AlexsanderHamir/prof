package testpaths

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ErrGoModNotFound is returned when no ancestor directory contains go.mod.
var ErrGoModNotFound = errors.New("go.mod not found")

// ModuleRoot returns the absolute path to the Go module root by walking up from
// the current working directory until go.mod is found.
func ModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return filepath.Abs(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrGoModNotFound
		}
		dir = parent
	}
}

// TestsAssetsDir returns <moduleRoot>/tests/assets.
func TestsAssetsDir() (string, error) {
	root, err := ModuleRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "tests", "assets"), nil
}

// Asset joins elem under tests/assets (relative to module root).
func Asset(elem ...string) (string, error) {
	base, err := TestsAssetsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{base}, elem...)...), nil
}

// MustAsset is like Asset but fails the test on error.
func MustAsset(t testing.TB, elem ...string) string {
	t.Helper()
	p, err := Asset(elem...)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

package tests

import (
	"path/filepath"
	"testing"
)

// edgecasesFixturesDir returns the absolute path to tests/assets/fixtures.
func edgecasesFixturesDir(t *testing.T) string {
	t.Helper()
	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("getProjectRoot: %v", err)
	}
	return filepath.Join(root, testDirName, fixturesSubdir)
}

// edgecasesFixturePath returns the absolute path to a committed fixture file
// under tests/assets/fixtures/ (e.g. fixtureCPUFile).
func edgecasesFixturePath(t *testing.T, fileName string) string {
	t.Helper()
	return filepath.Join(edgecasesFixturesDir(t), fileName)
}

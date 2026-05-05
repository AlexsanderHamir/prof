package tests

import (
	_ "embed" // embedded assets for synthetic test modules (see go:embed below)
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

// integrationEnvDir is the per-scenario directory name under tests/ (package cwd).
func integrationEnvDir(label string) string {
	return envDirNameStatic + " " + label
}

//go:embed assets/utils.go.txt
var utilsTemplate string

// BenchmarkContent is embedded benchmark test source used when creating synthetic modules.
//
//go:embed assets/benchmark_test.go.txt
var BenchmarkContent string

func createPackage(dir string) error {
	utilsDir := filepath.Join(dir, "utils")
	if err := os.MkdirAll(utilsDir, internal.PermDir); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	utilsPath := filepath.Join(utilsDir, "utils.go")
	return os.WriteFile(utilsPath, []byte(utilsTemplate), internal.PermFile)
}

func createBenchmarkFile(dir string) error {
	benchPath := filepath.Join(dir, "benchmark_test.go")
	return os.WriteFile(benchPath, []byte(BenchmarkContent), internal.PermFile)
}

func getProjectRoot() (string, error) {
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
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}

func setupEnviroment(t *testing.T, envDir string) {
	t.Helper()
	if err := os.Mkdir(envDir, internal.PermDir); err != nil && !os.IsExist(err) {
		t.Fatalf("couldn't create environment dir: %v", err)
	}

	goModPath := filepath.Join(envDir, "go.mod")
	if _, statErr := os.Stat(goModPath); statErr != nil {
		if !os.IsNotExist(statErr) {
			t.Fatalf("stat go.mod: %v", statErr)
		}
		cmd := exec.Command("go", "mod", "init", moduleName)
		cmd.Dir = envDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to initialize Go module: %v\nOutput: %s", err, output)
		}
	}

	if err := createPackage(envDir); err != nil {
		t.Fatalf("failed to create package: %v", err)
	}

	if err := createBenchmarkFile(envDir); err != nil {
		t.Fatalf("failed to create benchmark file: %v", err)
	}
}

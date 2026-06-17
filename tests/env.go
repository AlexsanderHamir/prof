package tests

import (
	_ "embed" // embedded assets for synthetic test modules (see go:embed below)
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/workspace"
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
	if err := os.MkdirAll(utilsDir, workspace.PermDir); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	utilsPath := filepath.Join(utilsDir, "utils.go")
	return os.WriteFile(utilsPath, []byte(utilsTemplate), workspace.PermFile)
}

func createBenchmarkFile(dir string) error {
	benchPath := filepath.Join(dir, "benchmark_test.go")
	return os.WriteFile(benchPath, []byte(BenchmarkContent), workspace.PermFile)
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

// materializeSyntheticEnv writes the synthetic Go module + benchmark file to
// envDir. Idempotent: skips `go mod init` if go.mod already exists, so it is
// safe to call from both ensureSharedEnv (sync.Once) and setupEnviroment.
func materializeSyntheticEnv(envDir string) error {
	if err := os.Mkdir(envDir, workspace.PermDir); err != nil && !os.IsExist(err) {
		return fmt.Errorf("create environment dir: %w", err)
	}

	goModPath := filepath.Join(envDir, "go.mod")
	if _, statErr := os.Stat(goModPath); statErr != nil {
		if !os.IsNotExist(statErr) {
			return fmt.Errorf("stat go.mod: %w", statErr)
		}
		cmd := exec.Command("go", "mod", "init", moduleName)
		cmd.Dir = envDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod init: %w (output: %s)", err, output)
		}
	}

	if err := createPackage(envDir); err != nil {
		return fmt.Errorf("create package: %w", err)
	}

	if err := createBenchmarkFile(envDir); err != nil {
		return fmt.Errorf("create benchmark file: %w", err)
	}

	return nil
}

func setupEnviroment(t *testing.T, envDir string) {
	t.Helper()
	if err := materializeSyntheticEnv(envDir); err != nil {
		t.Fatalf("setup synthetic env: %v", err)
	}
}

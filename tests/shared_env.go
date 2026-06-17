package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Shared synthetic Go module reused by every subprocess-based scenario that
// doesn't care about per-test env state (TestAutoEndToEnd,
// TestProfileValidation, TestCommandValidation). The module is materialized
// at most once per `go test` process and torn down by TestMain on exit.
//
// `prof auto` always cleans bench/<tag>/ before each run (see internal/workspace),
// so reusing the same tag across scenarios is safe — leftover artifacts from a
// previous test never leak into the next test's assertions.
var (
	sharedEnvOnce sync.Once
	sharedEnvDir  string
	errSharedEnv  error
)

// ensureSharedEnv builds the synthetic env once and returns its absolute path.
// Pays the ~1s setup cost once per `go test` process instead of per scenario.
func ensureSharedEnv(t *testing.T) string {
	t.Helper()

	sharedEnvOnce.Do(func() {
		root, err := getProjectRoot()
		if err != nil {
			errSharedEnv = fmt.Errorf("getProjectRoot: %w", err)
			return
		}

		envDir := integrationEnvDir(sharedEnvLabel)
		envFullPath := filepath.Join(root, testDirName, envDir)

		if err = materializeSyntheticEnv(envFullPath); err != nil {
			errSharedEnv = err
			return
		}

		src, err := ensureCachedProfBinary(root)
		if err != nil {
			errSharedEnv = fmt.Errorf("prof cache build: %w", err)
			return
		}
		dst := filepath.Join(envFullPath, profBinaryName())
		if err = copyProfBinary(src, dst); err != nil {
			errSharedEnv = fmt.Errorf("copy prof binary: %w", err)
			return
		}

		sharedEnvDir = envFullPath
	})

	if errSharedEnv != nil {
		t.Fatalf("shared env: %v", errSharedEnv)
	}
	return sharedEnvDir
}

func cleanupSharedEnv() {
	if sharedEnvDir == "" {
		return
	}
	_ = os.RemoveAll(sharedEnvDir)
}

// TestMain runs all package tests and cleans up the shared synthetic env on
// exit. Tests that need their own scratch env (TestManualCommand,
// TestTrackerBasicRun) continue to manage their own t.Cleanup blocks.
func TestMain(m *testing.M) {
	code := m.Run()
	cleanupSharedEnv()
	os.Exit(code)
}

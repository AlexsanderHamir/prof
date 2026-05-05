package tests

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

// profCacheDir is a single build output reused for all integration scenarios in one `go test` run.
const profCacheDir = ".integration_prof_build"

var (
	integrationProfOnce     sync.Once
	integrationProfCached   string
	errIntegrationProfBuild error
)

// profBinaryName is the built prof executable filename. On Windows, os/exec
// requires a recognized extension (e.g. .exe); a bare "prof" fails lookup.
func profBinaryName() string {
	if runtime.GOOS == "windows" {
		return "prof.exe"
	}
	return "prof"
}

func copyProfBinary(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if st, statErr := os.Stat(src); statErr == nil {
			_ = os.Chmod(dst, st.Mode()&0o777)
		}
	}
	return nil
}

// ensureCachedProfBinary builds cmd/prof once per test process; callers copy into each env dir.
func ensureCachedProfBinary(projectRoot string) (string, error) {
	integrationProfOnce.Do(func() {
		cacheDir := filepath.Join(projectRoot, testDirName, profCacheDir)
		if err := os.MkdirAll(cacheDir, internal.PermDir); err != nil {
			errIntegrationProfBuild = err
			return
		}
		out := filepath.Join(cacheDir, profBinaryName())
		cmdProfDir := filepath.Join(projectRoot, "cmd", "prof")
		buildCmd := exec.Command("go", "build", "-o", out, ".")
		buildCmd.Dir = cmdProfDir
		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			errIntegrationProfBuild = fmt.Errorf("failed to build prof binary: %w\nOutput: %s", err, buildOutput)
			return
		}
		integrationProfCached = out
	})
	return integrationProfCached, errIntegrationProfBuild
}

func setUpProf(t *testing.T, projectRoot, envDir string) {
	t.Helper()

	dst := filepath.Join(projectRoot, testDirName, envDir, profBinaryName())
	src, err := ensureCachedProfBinary(projectRoot)
	if err != nil {
		t.Fatalf("prof cache build: %v", err)
	}
	if err = copyProfBinary(src, dst); err != nil {
		t.Fatalf("failed to copy prof binary: %v", err)
	}
}

func buildProf(t *testing.T, outputPath, root string) {
	t.Helper()
	src, err := ensureCachedProfBinary(root)
	if err != nil {
		t.Fatalf("prof cache build: %v", err)
	}
	if err = copyProfBinary(src, outputPath); err != nil {
		t.Fatalf("failed to copy prof binary: %v", err)
	}
}

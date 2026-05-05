//go:build ignore

// fixtures_regen.go rebuilds tests/assets/fixtures/*.out from the synthetic
// benchmark environment described by tests/assets/{utils,benchmark_test}.go.txt.
//
// Run via:
//
//	go generate ./tests/...
//
// The tool:
//  1. Materializes a temporary Go module ("test-environment") in a temp dir.
//  2. Runs `go test -bench=^BenchmarkStringProcessor$ -count=1 -benchtime=2s
//     -cpuprofile=cpu.out -memprofile=memory.out`.
//  3. Copies the resulting binaries into tests/assets/fixtures/, replacing
//     whatever was there before.
//
// Regenerate whenever tests/assets/utils.go.txt or
// tests/assets/benchmark_test.go.txt changes. Committed binaries make the
// in-process filter tests deterministic and instantaneous.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	moduleName    = "test-environment"
	benchName     = "BenchmarkStringProcessor"
	benchtime     = "2s"
	utilsTmplName = "utils.go.txt"
	benchTmplName = "benchmark_test.go.txt"
	cpuOutName    = "cpu.out"
	memOutName    = "memory.out"

	assetsDir   = "assets"
	fixturesDir = "fixtures"

	permDir  = 0o755
	permFile = 0o644
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}

	utilsTmplPath := filepath.Join(cwd, assetsDir, utilsTmplName)
	benchTmplPath := filepath.Join(cwd, assetsDir, benchTmplName)
	fixturesPath := filepath.Join(cwd, assetsDir, fixturesDir)

	utilsContent, err := os.ReadFile(utilsTmplPath)
	if err != nil {
		log.Fatalf("read %s: %v", utilsTmplPath, err)
	}
	benchContent, err := os.ReadFile(benchTmplPath)
	if err != nil {
		log.Fatalf("read %s: %v", benchTmplPath, err)
	}

	tmpDir, err := os.MkdirTemp("", "prof-fixtures-*")
	if err != nil {
		log.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err = materializeEnv(tmpDir, utilsContent, benchContent); err != nil {
		log.Fatalf("materialize env: %v", err)
	}

	if err = runBench(tmpDir); err != nil {
		log.Fatalf("run bench: %v", err)
	}

	if err = os.MkdirAll(fixturesPath, permDir); err != nil {
		log.Fatalf("mkdir fixtures: %v", err)
	}

	for _, name := range []string{cpuOutName, memOutName} {
		src := filepath.Join(tmpDir, name)
		dst := filepath.Join(fixturesPath, fmt.Sprintf("%s_%s", benchName, name))
		if err = copyFile(src, dst); err != nil {
			log.Fatalf("copy %s: %v", name, err)
		}
		fmt.Printf("wrote %s\n", dst)
	}
}

func materializeEnv(dir string, utilsSrc, benchSrc []byte) error {
	utilsSubdir := filepath.Join(dir, "utils")
	if err := os.MkdirAll(utilsSubdir, permDir); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(utilsSubdir, "utils.go"), utilsSrc, permFile); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "benchmark_test.go"), benchSrc, permFile); err != nil {
		return err
	}
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod init: %w (output: %s)", err, out)
	}
	return nil
}

func runBench(dir string) error {
	cmd := exec.Command(
		"go", "test",
		"-run=^$",
		"-bench=^"+benchName+"$",
		"-benchmem",
		"-count=1",
		"-benchtime="+benchtime,
		"-cpuprofile="+cpuOutName,
		"-memprofile="+memOutName,
	)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, permFile)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

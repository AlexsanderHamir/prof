package benchmark

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/internal"
)

func findBenchmarkPackageDir(moduleRoot, benchmarkName string) (string, error) {
	pattern := regexp.MustCompile(`(?m)^\s*func\s+` + regexp.QuoteMeta(benchmarkName) + `\s*\(\s*b\s*\*\s*testing\.B\s*\)\s*{`)

	var foundDir string
	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if pattern.Find(data) != nil {
			foundDir = filepath.Dir(path)
			return nil
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if foundDir == "" {
		return "", fmt.Errorf("benchmark %s not found in module", benchmarkName)
	}
	return foundDir, nil
}

func buildBenchmarkCommand(benchmarkName string, profiles []string, count int) ([]string, error) {
	cmd := []string{
		"go", "test", "-run=^$",
		fmt.Sprintf("-bench=^%s$", benchmarkName),
		"-benchmem",
		fmt.Sprintf("-count=%d", count),
	}
	for _, profile := range profiles {
		flag, exists := ProfileFlags[profile]
		if !exists {
			return nil, fmt.Errorf("profile %s is not supported", profile)
		}
		cmd = append(cmd, flag)
	}
	return cmd, nil
}

func getOutputDirectoriesPath(benchmarkName, tag string) (textDir string, binDir string) {
	tagDir := filepath.Join(internal.MainDirOutput, tag)
	textDir = filepath.Join(tagDir, internal.ProfileTextDir, benchmarkName)
	binDir = filepath.Join(tagDir, internal.ProfileBinDir, benchmarkName)
	return textDir, binDir
}

func runBenchmarkCommand(cmd []string, outputFile string, rootDir string) error {
	// #nosec G204 -- argv from buildBenchmarkCommand
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	if rootDir != "" {
		execCmd.Dir = rootDir
	}
	output, err := execCmd.CombinedOutput()
	fmt.Println("🚀 ==================== BENCHMARK OUTPUT ==================== 🚀")
	fmt.Println(string(output))
	fmt.Println("📊 ========================================================== 📊")
	if err != nil {
		if strings.Contains(string(output), moduleNotFoundMsg) {
			return fmt.Errorf("❌ %s - ensure you're in a Go project directory 📁", moduleNotFoundMsg)
		}
		return fmt.Errorf("💥 BENCHMARK COMMAND FAILED 💥\n%s", string(output))
	}
	return os.WriteFile(outputFile, output, internal.PermFile)
}

func runBenchmark(benchmarkName string, profiles []string, count int, tag string) error {
	cmd, err := buildBenchmarkCommand(benchmarkName, profiles, count)
	if err != nil {
		return err
	}
	textDir, binDir := getOutputDirectoriesPath(benchmarkName, tag)
	moduleRoot, err := internal.FindGoModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to find Go module root: %w", err)
	}
	pkgDir, err := findBenchmarkPackageDir(moduleRoot, benchmarkName)
	if err != nil {
		return fmt.Errorf("failed to locate benchmark %s: %w", benchmarkName, err)
	}
	outputFile := filepath.Join(textDir, fmt.Sprintf("%s.%s", benchmarkName, internal.TextExtension))
	if err = runBenchmarkCommand(cmd, outputFile, pkgDir); err != nil {
		return err
	}
	if err = moveProfileFiles(benchmarkName, profiles, pkgDir, binDir); err != nil {
		return err
	}
	return moveTestFiles(benchmarkName, pkgDir, binDir)
}

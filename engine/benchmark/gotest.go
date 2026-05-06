package benchmark

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal"
)

func findBenchmarkPackageDir(moduleRoot, benchmarkName string) (string, error) {
	pattern := regexp.MustCompile(`(?m)^\s*func\s+` + regexp.QuoteMeta(benchmarkName) + `\s*\(\s*b\s*\*\s*testing\.B\s*\)\s*{`)

	var foundDir string
	err := walkTestGoFiles(moduleRoot, func(path string, data []byte) error {
		if pattern.Find(data) != nil {
			foundDir = filepath.Dir(path)
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
		internal.GoBinaryName, internal.GoTestSubcommand, "-run=^$",
		fmt.Sprintf("-bench=^%s$", benchmarkName),
		"-benchmem",
		fmt.Sprintf("-count=%d", count),
	}
	flags, err := benchmarkCatalog.GoTestProfileArgs(profiles)
	if err != nil {
		return nil, err
	}
	return append(cmd, flags...), nil
}

func getOutputDirectoriesPath(benchmarkName, tag string) (textDir string, binDir string) {
	tagDir := filepath.Join(internal.MainDirOutput, tag)
	textDir = filepath.Join(tagDir, internal.ProfileTextDir, benchmarkName)
	binDir = filepath.Join(tagDir, internal.ProfileBinDir, benchmarkName)
	return textDir, binDir
}

func runBenchmarkCommand(runner tooling.Runner, cmd []string, outputFile string, rootDir string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	ctx := context.Background()
	output, err := runner.Run(ctx, cmd, tooling.RunOpts{Dir: rootDir, Combined: true})
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

func runBenchmark(runner tooling.Runner, benchmarkName string, profiles []string, count int, tag string) error {
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
	if err = runBenchmarkCommand(runner, cmd, outputFile, pkgDir); err != nil {
		return err
	}
	if err = moveProfileFiles(benchmarkName, profiles, pkgDir, binDir); err != nil {
		return err
	}
	return moveTestFiles(benchmarkName, pkgDir, binDir)
}

package collect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func findBenchmarkPackageDir(moduleRoot, benchmarkName string) (string, error) {
	pattern := regexp.MustCompile(`(?m)^\s*func\s+` + regexp.QuoteMeta(benchmarkName) + `\s*\(\s*b\s*\*\s*testing\.B\s*\)\s*{`)

	var foundDir string
	err := walkTestGoFiles(moduleRoot, moduleRoot, func(path string, data []byte) error {
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
		workspace.GoBinaryName, workspace.GoTestSubcommand, "-run=^$",
		fmt.Sprintf("-bench=^%s$", benchmarkName),
		"-benchmem",
		fmt.Sprintf("-count=%d", count),
	}
	flags, err := profileCatalog.GoTestProfileArgs(profiles)
	if err != nil {
		return nil, err
	}
	return append(cmd, flags...), nil
}

func runBenchmarkCommand(runner tooling.Runner, cmd []string, outputFile string, rootDir string, progress termui.Progress) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	ctx := context.Background()
	var output []byte
	err := termui.RunWhile(os.Stderr, int(os.Stderr.Fd()), progress, func() error {
		var runErr error
		output, runErr = runner.Run(ctx, cmd, tooling.RunOpts{Dir: rootDir, Combined: true})
		return runErr
	})
	if err != nil {
		if strings.Contains(string(output), moduleNotFoundMsg) {
			return fmt.Errorf("%s - ensure you're in a Go project directory", moduleNotFoundMsg)
		}
		return fmt.Errorf("benchmark command failed:\n%s", string(output))
	}
	if writeErr := os.WriteFile(outputFile, output, workspace.PermFile); writeErr != nil {
		return writeErr
	}
	termui.DoneLine(os.Stderr, int(os.Stderr.Fd()), progress)
	return nil
}

func runBenchmark(runner tooling.Runner, benchmarkName string, profiles []string, count int, tag string, progress termui.Progress) error {
	cmd, err := buildBenchmarkCommand(benchmarkName, profiles, count)
	if err != nil {
		return err
	}
	layout, err := workspace.TagLayoutFromCWD(tag)
	if err != nil {
		return err
	}
	moduleRoot, err := workspace.FindModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to find Go module root: %w", err)
	}
	pkgDir, err := findBenchmarkPackageDir(moduleRoot, benchmarkName)
	if err != nil {
		return fmt.Errorf("failed to locate benchmark %s: %w", benchmarkName, err)
	}
	outputFile := layout.BenchText(benchmarkName)
	binDir := filepath.Join(layout.Root, workspace.ProfileBinDir, benchmarkName)
	if err = runBenchmarkCommand(runner, cmd, outputFile, pkgDir, progress); err != nil {
		return err
	}
	if err = moveProfileFiles(benchmarkName, profiles, pkgDir, binDir); err != nil {
		return err
	}
	return moveTestFiles(benchmarkName, pkgDir, binDir)
}

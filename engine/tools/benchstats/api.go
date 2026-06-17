package benchstats

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// RunBenchStats runs benchstat on two collected benchmark text files.
func RunBenchStats(runner tooling.Runner, baseTag, currentTag, benchName string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	benchDir := workspace.MainDirOutput
	if _, err := os.Stat(benchDir); os.IsNotExist(err) {
		return errors.New("bench directory not found in current directory")
	}

	baseTagPath := filepath.Join(benchDir, baseTag)
	currentTagPath := filepath.Join(benchDir, currentTag)

	if _, err := os.Stat(baseTagPath); os.IsNotExist(err) {
		return fmt.Errorf("base tag directory '%s' not found in bench directory", baseTag)
	}
	if _, err := os.Stat(currentTagPath); os.IsNotExist(err) {
		return fmt.Errorf("current tag directory '%s' not found in bench directory", currentTag)
	}

	baseLayout := workspace.TagLayout{Tag: baseTag, Root: baseTagPath}
	curLayout := workspace.TagLayout{Tag: currentTag, Root: currentTagPath}
	baseTextPath := baseLayout.BenchText(benchName)
	currentTextPath := curLayout.BenchText(benchName)

	if _, err := os.Stat(baseTextPath); os.IsNotExist(err) {
		return fmt.Errorf("base benchmark text file not found: %s", baseTextPath)
	}
	if _, err := os.Stat(currentTextPath); os.IsNotExist(err) {
		return fmt.Errorf("current benchmark text file not found: %s", currentTextPath)
	}

	if _, err := tooling.LookPath(workspace.ToolNameBenchstat); err != nil {
		return errors.New("benchstat command not found. Please install it first: go install golang.org/x/perf/cmd/benchstat@latest")
	}

	output, runErr := runner.Run(context.Background(), []string{workspace.ToolNameBenchstat, baseTextPath, currentTextPath}, tooling.RunOpts{Combined: true})
	if runErr != nil {
		return fmt.Errorf("failed to run benchstat: %w, output: %s", runErr, string(output))
	}

	fmt.Println("Benchmark comparison results:")
	fmt.Println(string(output))

	resultsDir := filepath.Join(benchDir, workspace.ToolDir, workspace.ToolNameBenchstat)
	if mkErr := os.MkdirAll(resultsDir, workspace.PermDir); mkErr != nil {
		return fmt.Errorf("failed to create results directory: %w", mkErr)
	}

	resultsFile := filepath.Join(resultsDir, benchName+workspace.ToolsResultsSuffix)
	if werr := os.WriteFile(resultsFile, output, workspace.PermFile); werr != nil {
		return fmt.Errorf("failed to save results to file: %w", werr)
	}

	fmt.Printf("Results saved to: %s\n", resultsFile)
	return nil
}

package benchstats

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
)

const (
	benchstatCommand = "benchstat"
)

func RunBenchStats(baseTag, currentTag, benchName string) error {
	// 1. Look for the bench directory created by our library under the current directory where this command will be run
	benchDir := internal.MainDirOutput
	if _, err := os.Stat(benchDir); os.IsNotExist(err) {
		return errors.New("bench directory not found in current directory")
	}

	// 2. Inside the bench directory, look for the passed tags, if one of them don't exist return an error.
	baseTagPath := filepath.Join(benchDir, baseTag)
	currentTagPath := filepath.Join(benchDir, currentTag)

	if _, err := os.Stat(baseTagPath); os.IsNotExist(err) {
		return fmt.Errorf("base tag directory '%s' not found in bench directory", baseTag)
	}

	if _, err := os.Stat(currentTagPath); os.IsNotExist(err) {
		return fmt.Errorf("current tag directory '%s' not found in bench directory", currentTag)
	}

	// 3. Once both directories are found look for the path text/{benchmarkname}/{benchmarkname}.txt, this file contains the data for benchstats
	baseTextPath := filepath.Join(baseTagPath, internal.ProfileTextDir, benchName, benchName+"."+internal.TextExtension)
	currentTextPath := filepath.Join(currentTagPath, internal.ProfileTextDir, benchName, benchName+"."+internal.TextExtension)

	if _, err := os.Stat(baseTextPath); os.IsNotExist(err) {
		return fmt.Errorf("base benchmark text file not found: %s", baseTextPath)
	}

	if _, err := os.Stat(currentTextPath); os.IsNotExist(err) {
		return fmt.Errorf("current benchmark text file not found: %s", currentTextPath)
	}

	// 4. Run benchstats programmatically, if benchstats is not installed on the machine return an error.
	// Check if benchstats is installed
	if _, err := exec.LookPath(benchstatCommand); err != nil {
		return errors.New("benchstat command not found. Please install it first: go install golang.org/x/perf/cmd/benchstat@latest")
	}

	// Run benchstat command
	cmd := exec.Command(benchstatCommand, baseTextPath, currentTextPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run benchstat: %w, output: %s", err, string(output))
	}

	// 5. Print the output to the terminal and save it under bench/tools/benchstats/{benchmarkname}_results.txt
	fmt.Println("Benchmark comparison results:")
	fmt.Println(string(output))

	// Save results to file
	resultsDir := filepath.Join(benchDir, internal.ToolDir, benchstatCommand)
	if err = os.MkdirAll(resultsDir, internal.PermDir); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	resultsFile := filepath.Join(resultsDir, benchName+internal.ToolsResultsSuffix)
	if err = os.WriteFile(resultsFile, output, internal.PermFile); err != nil {
		return fmt.Errorf("failed to save results to file: %w", err)
	}

	fmt.Printf("Results saved to: %s\n", resultsFile)
	return nil
}

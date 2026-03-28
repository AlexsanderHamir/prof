package collector

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
)

// GetProfileTextOutput runs go tool pprof text listing and writes output to outputFile.
func GetProfileTextOutput(binaryFile, outputFile string) error {
	cmd := append([]string{"go", "tool", "pprof"}, pprofTextListArgs()...)
	cmd = append(cmd, binaryFile)
	// #nosec G204 -- argv built here, binary path only
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof command failed: %w", err)
	}
	return os.WriteFile(outputFile, out, internal.PermFile)
}

// GetPNGOutput renders a PNG flame-style view via go tool pprof -png.
func GetPNGOutput(binaryFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", "-png", binaryFile}
	// #nosec G204
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof PNG generation failed: %w", err)
	}
	return os.WriteFile(outputFile, out, internal.PermFile)
}

func writeFunctionListPprof(function, binaryFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", fmt.Sprintf("-list=%s", function), binaryFile}
	// #nosec G204
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof list command failed: %w", err)
	}
	if err = os.WriteFile(outputFile, out, internal.PermFile); err != nil {
		return fmt.Errorf("write function content: %w", err)
	}
	slog.Info("Collected function", "function", function)
	return nil
}

// GetFunctionsOutput runs pprof -list for each function name into basePath.
func GetFunctionsOutput(functions []string, binaryPath, basePath string) error {
	for _, name := range functions {
		out := filepath.Join(basePath, name+"."+internal.TextExtension)
		if err := writeFunctionListPprof(name, binaryPath, out); err != nil {
			return fmt.Errorf("extract function %s: %w", name, err)
		}
	}
	return nil
}

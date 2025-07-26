package collector

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
)

func ensureDirExists(basePath string) error {
	_, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(basePath, shared.PermDir)
		}
		return err
	}

	return nil
}

// getFunctionPprofContent gets code line level mapping of specified function
// and writes the data to a file named after the function.
func getFunctionPprofContent(function, binaryFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", fmt.Sprintf("-list=%s", function), binaryFile}

	// #nosec ProfileTextDir04 -- cmd is constructed internally by getFunctionPprofContent(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof list command failed: %w", err)
	}

	if err = os.WriteFile(outputFile, output, shared.PermFile); err != nil {
		return fmt.Errorf("failed to write function content: %w", err)
	}

	slog.Info("Collected function", "function", function)
	return nil
}

func collectFunctions(outputTextFilePath, profileDirPath, fullBinaryPath string, functionFilter config.FunctionFilter) error {
	var functions []string
	functions, err := parser.GetAllFunctionNames(outputTextFilePath, functionFilter)
	if err != nil {
		return fmt.Errorf("failed to extract function names: %w", err)
	}

	functionDir := path.Join(profileDirPath, "functions")
	if err = ensureDirExists(functionDir); err != nil {
		return err
	}

	if err = SaveAllFunctionsPprofContents(functions, fullBinaryPath, functionDir); err != nil {
		return fmt.Errorf("getAllFunctionsPprofContents failed: %w", err)
	}

	return nil
}

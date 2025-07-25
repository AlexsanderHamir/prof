package collector

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
)

const globalSign = "*"

// RunCollector handles data organization without wrapping go test.
func RunCollector(files []string, tag string) error {
	if err := ensureDirExists(shared.MainDirOutput); err != nil {
		return err
	}

	tagDir := filepath.Join(shared.MainDirOutput, tag)
	err := shared.CleanOrCreateDir(tagDir)
	if err != nil {
		return fmt.Errorf("CleanOrCreateDir failed: %w", err)
	}

	cfg, err := config.LoadFromFile(shared.ConfigFilename)
	if err != nil {
		cfg = &config.Config{}
	}

	var functionFilter config.FunctionFilter

	globalFilter, hasGlobalFilter := cfg.FunctionFilter[globalSign]
	if hasGlobalFilter {
		functionFilter = globalFilter
	}

	for _, binaryFilePath := range files {
		binaryDirName := filepath.Base(binaryFilePath)
		fileName := strings.TrimSuffix(binaryDirName, filepath.Ext(binaryDirName))
		profileDirPath := path.Join(tagDir, fileName)
		if err = ensureDirExists(profileDirPath); err != nil {
			return err
		}

		if !hasGlobalFilter {
			localFilter, hasLocalFilter := cfg.FunctionFilter[fileName]
			if hasLocalFilter {
				functionFilter = localFilter
			}
		}

		outputTextFilePath := path.Join(profileDirPath, fileName+"."+shared.TextExtension)
		if err = GenerateProfileTextOutput(binaryFilePath, outputTextFilePath); err != nil {
			return err
		}

		var functions []string
		functions, err = parser.GetAllFunctionNames(outputTextFilePath, functionFilter)
		if err != nil {
			return fmt.Errorf("failed to extract function names: %w", err)
		}

		functionDir := path.Join(profileDirPath, "functions")
		if err = ensureDirExists(functionDir); err != nil {
			return err
		}

		if err = SaveAllFunctionsPprofContents(functions, binaryFilePath, functionDir); err != nil {
			return fmt.Errorf("getAllFunctionsPprofContents failed: %w", err)
		}
	}
	return nil
}

func GenerateProfileTextOutput(binaryFile, outputFile string) error {
	pprofTextParams := getPprofTextParams()
	cmd := append([]string{"go", "tool", "pprof"}, pprofTextParams...)
	cmd = append(cmd, binaryFile)

	// #nosec G204 -- cmd is constructed internally by generateTextProfile(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof command failed: %w", err)
	}

	return os.WriteFile(outputFile, output, shared.PermFile)
}

func GeneratePNGVisualization(binaryFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", "-png", binaryFile}

	// #nosec G204 -- cmd is constructed internally by generatePNGVisualization(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof PNG generation failed: %w", err)
	}

	return os.WriteFile(outputFile, output, shared.PermFile)
}

// GetFunctionPprofContent gets code line level mapping of specified function
// and writes the data to a file named after the function.
func GetFunctionPprofContent(function, binaryFile, outputFile string) error {
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

// SaveAllFunctionsPprofContents calls [GetFunctionPprofContent] sequentially.
func SaveAllFunctionsPprofContents(functions []string, binaryPath, basePath string) error {
	for _, functionName := range functions {
		outputFile := filepath.Join(basePath, functionName+"."+shared.TextExtension)
		if err := GetFunctionPprofContent(functionName, binaryPath, outputFile); err != nil {
			return fmt.Errorf("failed to extract function content for %s: %w", functionName, err)
		}
	}

	return nil
}

package collector

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/config"
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

	var profileDirPath string
	for _, fullBinaryPath := range files {
		fileName := getFileName(fullBinaryPath)
		profileDirPath, err = createProfileDirectory(tagDir, fileName)
		if err != nil {
			return fmt.Errorf("createProfileDirectory failed: %w", err)
		}

		if !hasGlobalFilter {
			localFilter, hasLocalFilter := cfg.FunctionFilter[fileName]
			if hasLocalFilter {
				functionFilter = localFilter
			}
		}

		outputTextFilePath := path.Join(profileDirPath, fileName+"."+shared.TextExtension)
		if err = GenerateProfileTextOutput(fullBinaryPath, outputTextFilePath); err != nil {
			return err
		}

		if err = collectFunctions(outputTextFilePath, profileDirPath, fullBinaryPath, functionFilter); err != nil {
			return fmt.Errorf("collectFunctions failed: %w", err)
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

// SaveAllFunctionsPprofContents calls [GetFunctionPprofContent] sequentially.
func SaveAllFunctionsPprofContents(functions []string, binaryPath, basePath string) error {
	for _, functionName := range functions {
		outputFile := filepath.Join(basePath, functionName+"."+shared.TextExtension)
		if err := getFunctionPprofContent(functionName, binaryPath, outputFile); err != nil {
			return fmt.Errorf("failed to extract function content for %s: %w", functionName, err)
		}
	}

	return nil
}

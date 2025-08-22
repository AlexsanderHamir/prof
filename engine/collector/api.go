package collector

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
)

// RunCollector handles data organization without wrapping go test.
func RunCollector(files []string, tag string) error {
	if err := ensureDirExists(internal.MainDirOutput); err != nil {
		return err
	}

	tagDir := filepath.Join(internal.MainDirOutput, tag)
	err := internal.CleanOrCreateTag(tagDir)
	if err != nil {
		return fmt.Errorf("CleanOrCreateTag failed: %w", err)
	}

	cfg, err := internal.LoadFromFile(internal.ConfigFilename)
	if err != nil {
		cfg = &internal.Config{}
	}

	var functionFilter internal.FunctionFilter
	globalFilter, hasGlobalFilter := cfg.FunctionFilter[internal.GlobalSign]
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
			functionFilter = internal.FunctionFilter{} // clean previous one
			localFilter, hasLocalFilter := cfg.FunctionFilter[fileName]
			if hasLocalFilter {
				functionFilter = localFilter
			}
		}

		outputTextFilePath := path.Join(profileDirPath, fileName+"."+internal.TextExtension)
		if err = GetProfileTextOutput(fullBinaryPath, outputTextFilePath); err != nil {
			return err
		}

		if err = collectFunctions(profileDirPath, fullBinaryPath, functionFilter); err != nil {
			return fmt.Errorf("collectFunctions failed: %w", err)
		}
	}
	return nil
}

func GetProfileTextOutput(binaryFile, outputFile string) error {
	pprofTextParams := getPprofTextParams()
	cmd := append([]string{"go", "tool", "pprof"}, pprofTextParams...)
	cmd = append(cmd, binaryFile)

	// #nosec G204 -- cmd is constructed internally by generateTextProfile(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof command failed: %w", err)
	}

	return os.WriteFile(outputFile, output, internal.PermFile)
}

func GetPNGOutput(binaryFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", "-png", binaryFile}

	// #nosec G204 -- cmd is constructed internally by GetPNGOutput(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof PNG generation failed: %w", err)
	}

	return os.WriteFile(outputFile, output, internal.PermFile)
}

// GetFunctionsOutput calls [GetFunctionPprofContent] sequentially.
func GetFunctionsOutput(functions []string, binaryPath, basePath string) error {
	for _, functionName := range functions {
		outputFile := filepath.Join(basePath, functionName+"."+internal.TextExtension)
		if err := getFunctionPprofContent(functionName, binaryPath, outputFile); err != nil {
			return fmt.Errorf("failed to extract function content for %s: %w", functionName, err)
		}
	}

	return nil
}

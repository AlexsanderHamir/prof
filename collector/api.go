package collector

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/shared"
)

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

	for _, binaryFilePath := range files {
		// 3. create a dir for each profile file
		fileName := strings.TrimSuffix(binaryFilePath, filepath.Ext(binaryFilePath))
		profilepath := path.Join(tagDir, fileName)
		if err := ensureDirExists(profilepath); err != nil {
			return err
		}

		// 4. Collect results.
		// 1. generate txt files
		outputFilePath := path.Join(profilepath, fileName+"."+shared.TextExtension)
		if err := GenerateProfileTextOutput(binaryFilePath, outputFilePath); err != nil {
			return err
		}

		// 2. collect functions according to config

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

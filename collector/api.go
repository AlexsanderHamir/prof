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

// Random Thoughts

// ## Actions
// 1. CLI
// Don't specify profiles

// 2. How are we going to organize it?
// branch/Tag
// 	profile_name/
//	profile_name.txt
// 		profile_functions/
//			Get.txt
//			Put.txt

// 3. Collect
// use config

// runCollecor handles data organization without wrapping go test.
func RunCollector(files []string, tag string) error {
	if err := ensureDirExists(shared.MainDirOutput); err != nil {
		return err
	}

	tagDir := filepath.Join(shared.MainDirOutput, tag)
	err := shared.CleanOrCreateDir(tagDir)
	if err != nil {
		return fmt.Errorf("CleanOrCreateDir failed: %w", err)
	}

	for _, file := range files {
		fileName := strings.TrimSuffix(file, filepath.Ext(file))
		profilepath := path.Join(tagDir, fileName)
		if err := ensureDirExists(profilepath); err != nil {
			return err
		}
		// 3. create a dir for each profile file
		// 4. Organize the results.
		fmt.Println("profile: ", file)
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

package benchmark

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexsanderHamir/prof/internal/shared"
)

func getProfileFlags() map[string]string {
	return map[string]string{
		"cpu":    "-cpuprofile=cpu.out",
		"memory": "-memprofile=memory.out",
		"mutex":  "-mutexprofile=mutex.out",
		"block":  "-blockprofile=block.out",
	}
}

const (
	binExtension        = "out"
	descriptionFileName = "description.txt"
	moduleNotFoundMsg   = "go: cannot find main module"
	waitForFiles        = 100
)

// createBenchDirectories creates the main structure of the library's output.
func createBenchDirectories(tagDir string, benchmarks []string) error {
	binDir := filepath.Join(tagDir, shared.ProfileBinDir)
	textDir := filepath.Join(tagDir, shared.ProfileTextDir)
	descFile := filepath.Join(tagDir, descriptionFileName)

	// Create main directories
	if err := os.MkdirAll(binDir, shared.PermDir); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	if err := os.MkdirAll(textDir, shared.PermDir); err != nil {
		return fmt.Errorf("failed to create text directory: %w", err)
	}

	// Create benchmark subdirectories
	for _, benchmark := range benchmarks {
		if err := os.MkdirAll(filepath.Join(binDir, benchmark), shared.PermDir); err != nil {
			return fmt.Errorf("failed to create bin subdirectory for %s: %w", benchmark, err)
		}
		if err := os.MkdirAll(filepath.Join(textDir, benchmark), shared.PermDir); err != nil {
			return fmt.Errorf("failed to create text subdirectory for %s: %w", benchmark, err)
		}
	}

	// Create description file
	if err := os.WriteFile(descFile, []byte(""), shared.PermFile); err != nil {
		return fmt.Errorf("failed to create description file: %w", err)
	}

	slog.Info("Created directory structure", "dir", tagDir)
	return nil
}

// createProfileFunctionDirectories creates the structure for the code line level data collection.
func createProfileFunctionDirectories(tagDir string, profiles, benchmarks []string) error {
	for _, profile := range profiles {
		profileDir := filepath.Join(tagDir, profile+shared.FunctionsDirSuffix)
		if err := os.MkdirAll(profileDir, shared.PermDir); err != nil {
			return fmt.Errorf("failed to create profile directory %s: %w", profileDir, err)
		}

		for _, benchmark := range benchmarks {
			benchmarkDir := filepath.Join(profileDir, benchmark)
			if err := os.MkdirAll(benchmarkDir, shared.PermDir); err != nil {
				return fmt.Errorf("failed to create benchmark directory %s: %w", benchmarkDir, err)
			}
		}
	}

	slog.Info("Created profile function directories")
	return nil
}

// buildBenchmarkCommand builds the command to run the benchmark.
func buildBenchmarkCommand(benchmarkName string, profiles []string, count int) ([]string, error) {
	cmd := []string{
		"go", "test", "-run=^$",
		fmt.Sprintf("-bench=^%s$", benchmarkName),
		"-benchmem",
		fmt.Sprintf("-count=%d", count),
	}

	profileFlags := getProfileFlags()
	for _, profile := range profiles {
		flag, exists := profileFlags[profile]

		if !exists {
			return nil, fmt.Errorf("profile %s is not supported", profile)
		}

		cmd = append(cmd, flag)
	}

	return cmd, nil
}

// getOutputDirectories gets or creates the output directories.
func getOutputDirectories(benchmarkName, tag string) (textDir string, binDir string) {
	tagDir := filepath.Join(shared.MainDirOutput, tag)
	textDir = filepath.Join(tagDir, shared.ProfileTextDir, benchmarkName)
	binDir = filepath.Join(tagDir, shared.ProfileBinDir, benchmarkName)

	return textDir, binDir
}

func runBenchmarkCommand(cmd []string, outputFile string) error {
	// cmd[0] = executable program (e.g., "go")
	// cmd[1:] = arguments to pass to the program (e.g., ["test", "-run=^$", "-bench=..."])
	// #nosec G204 -- cmd is constructed internally by buildBenchmarkCommand(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)

	output, err := execCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), moduleNotFoundMsg) {
			return fmt.Errorf("%s - ensure you're in a Go project directory", moduleNotFoundMsg)
		}
		return errors.New(string(output))
	}

	return os.WriteFile(outputFile, output, shared.PermFile)
}

func moveProfileFiles(benchmarkName string, profiles []string, binDir string) error {
	profileFlags := getProfileFlags()

	for _, profile := range profiles {
		flag, exists := profileFlags[profile]
		if !exists {
			continue
		}

		profileFile := strings.Split(flag, "=")[1]
		if _, err := os.Stat(profileFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("failed to stat profile file %s: %w", profileFile, err)
		}

		// Wait for file to be fully written (workaround)
		time.Sleep(waitForFiles * time.Millisecond)

		newPath := filepath.Join(binDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, binExtension))
		if err := os.Rename(profileFile, newPath); err != nil {
			return fmt.Errorf("failed to move profile file %s: %w", profileFile, err)
		}
	}
	return nil
}

func moveTestFiles(benchmarkName, binDir string) error {
	files, err := filepath.Glob("*.test")
	if err != nil {
		return err
	}

	for _, file := range files {
		newPath := filepath.Join(binDir, fmt.Sprintf("%s_%s", benchmarkName, filepath.Base(file)))
		if err = os.Rename(file, newPath); err != nil {
			return fmt.Errorf("failed to move test file %s: %w", file, err)
		}
	}
	return nil
}

// ProfilePaths holds paths for profile text, binary, and output directories.
type ProfilePaths struct {
	// Desired file path for specified profile
	ProfileTextFile string

	// Desired bin path for specified profile
	ProfileBinaryFile string

	// Desired benchmark directory for function data collection
	FunctionDirectory string
}

// getProfilePaths constructs file paths for benchmark profile data organized by tag and benchmark.
//
// Returns paths for:
//   - ProfileTextFile: bench/{tag}/text/{benchmarkName}/{benchmarkName}_{profile}.txt
//   - ProfileBinaryFile: bench/{tag}/bin/{benchmarkName}/{benchmarkName}_{profile}.out
//   - FunctionDirectory: bench/{tag}/{profile}_functions/{benchmarkName}/
//
// Example with tag="v1.0", benchmarkName="BenchmarkPool", profile="cpu":
//   - bench/v1.0/text/BenchmarkPool/BenchmarkPool_cpu.txt
//   - bench/v1.0/bin/BenchmarkPool/BenchmarkPool_cpu.out
//   - bench/v1.0/cpu_functions/BenchmarkPool/function1.txt
func getProfilePaths(tag, benchmarkName, profile string) ProfilePaths {
	tagDir := filepath.Join("bench", tag)
	profileTextFile := fmt.Sprintf("%s_%s.%s", benchmarkName, profile, shared.TextExtension)
	profileBinFile := fmt.Sprintf("%s_%s.%s", benchmarkName, profile, binExtension)

	return ProfilePaths{
		ProfileTextFile:   filepath.Join(tagDir, "text", benchmarkName, profileTextFile),
		ProfileBinaryFile: filepath.Join(tagDir, "bin", benchmarkName, profileBinFile),
		FunctionDirectory: filepath.Join(tagDir, profile+shared.FunctionsDirSuffix, benchmarkName),
	}
}

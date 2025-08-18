package benchmark

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// findBenchmarkPackageDir walks the module root to locate the package directory
// that defines the uniquely named benchmark function.
func findBenchmarkPackageDir(moduleRoot, benchmarkName string) (string, error) {
	pattern := regexp.MustCompile(`(?m)^\s*func\s+` + regexp.QuoteMeta(benchmarkName) + `\s*\(\s*b\s*\*\s*testing\.B\s*\)\s*{`)

	var foundDir string
	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			// skip common vendor-like directories
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if pattern.Find(data) != nil {
			foundDir = filepath.Dir(path)
			return errors.New("FOUND")
		}
		return nil
	})
	if err != nil && err.Error() != "FOUND" {
		return "", err
	}
	if foundDir == "" {
		return "", fmt.Errorf("benchmark %s not found in module", benchmarkName)
	}
	return foundDir, nil
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

func runBenchmarkCommand(cmd []string, outputFile string, rootDir string) error {
	// cmd[0] = executable program (e.g., "go")
	// cmd[1:] = arguments to pass to the program (e.g., ["test", "-run=^$", "-bench=..."])
	// #nosec G204 -- cmd is constructed internally by buildBenchmarkCommand(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	if rootDir != "" {
		execCmd.Dir = rootDir
	}

	output, err := execCmd.CombinedOutput()

	// Always print the output, even if there was an error - it may contain meaningful information
	fmt.Println("🚀 ==================== BENCHMARK OUTPUT ==================== 🚀")
	fmt.Println(string(output))
	fmt.Println("📊 ========================================================== 📊")

	if err != nil {
		if strings.Contains(string(output), moduleNotFoundMsg) {
			return fmt.Errorf("❌ %s - ensure you're in a Go project directory 📁", moduleNotFoundMsg)
		}
		return fmt.Errorf("💥 BENCHMARK COMMAND FAILED 💥\n%s", string(output))
	}

	return os.WriteFile(outputFile, output, shared.PermFile)
}

// profileFlagToFile extracts the file name from a profile flag like "-cpuprofile=cpu.out".
func profileFlagToFile(profile string, profileFlags map[string]string) (string, bool) {
	flag, exists := profileFlags[profile]
	if !exists {
		return "", false
	}
	expectedParts := 2
	parts := strings.SplitN(flag, "=", expectedParts)
	if len(parts) != expectedParts {
		return "", false
	}
	return parts[1], true
}

// findMostRecentFile searches for the most recently modified file named fileName under rootDir.
func findMostRecentFile(rootDir, fileName string) (string, error) {
	var latestPath string
	var latestMod time.Time

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != fileName {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return statErr
		}
		if info.ModTime().After(latestMod) {
			latestMod = info.ModTime()
			latestPath = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return latestPath, nil
}

// buildProfileDestPath builds the destination path for a profile binary output.
func buildProfileDestPath(binDir, benchmarkName, profile string) string {
	return filepath.Join(binDir, fmt.Sprintf("%s_%s.%s", benchmarkName, profile, binExtension))
}

// moveFileWithDelay waits for a short period and then renames the src to dst.
func moveFileWithDelay(src, dst string, delay time.Duration) error {
	time.Sleep(delay)
	return os.Rename(src, dst)
}

func moveProfileFiles(benchmarkName string, profiles []string, rootDir string, binDir string) error {
	profileFlags := getProfileFlags()

	for _, profile := range profiles {
		profileFile, ok := profileFlagToFile(profile, profileFlags)
		if !ok {
			continue
		}

		latestPath, err := findMostRecentFile(rootDir, profileFile)
		if err != nil {
			return fmt.Errorf("failed to search for profile files: %w", err)
		}
		if latestPath == "" {
			continue
		}

		destPath := buildProfileDestPath(binDir, benchmarkName, profile)
		if err = moveFileWithDelay(latestPath, destPath, waitForFiles*time.Millisecond); err != nil {
			return fmt.Errorf("failed to move profile file %s: %w", latestPath, err)
		}
	}
	return nil
}

func moveTestFiles(benchmarkName, rootDir, binDir string) error {
	var testFiles []string
	_ = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".test") {
			testFiles = append(testFiles, path)
		}
		return nil
	})

	for _, file := range testFiles {
		newPath := filepath.Join(binDir, fmt.Sprintf("%s_%s", benchmarkName, filepath.Base(file)))
		if err := os.Rename(file, newPath); err != nil {
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

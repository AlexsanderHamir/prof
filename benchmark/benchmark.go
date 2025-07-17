package benchmark

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/parser"
)

var profileFlags = map[string]string{
	"cpu":    "-cpuprofile=cpu.out",
	"memory": "-memprofile=memory.out",
	"mutex":  "-mutexprofile=mutex.out",
	"trace":  "-trace=trace.out",
}

var pprofTextParams = []string{
	"-nodecount=100000000",
	"-cum",
	"-edgefraction=0",
	"-nodefraction=0",
	"-top",
}

const (
	permDir  = 0o755
	permFile = 0o644
)

func SetupDirectories(tag string, benchmarks, profiles []string) error {
	if err := createBenchDirectories(tag, benchmarks); err != nil {
		return err
	}
	return createProfileFunctionDirectories(tag, profiles, benchmarks)
}

func createBenchDirectories(tag string, benchmarks []string) error {
	benchDir := "bench"
	tagDir := filepath.Join(benchDir, tag)
	binDir := filepath.Join(tagDir, "bin")
	textDir := filepath.Join(tagDir, "text")
	descFile := filepath.Join(tagDir, "description.txt")

	// Create main directories
	if err := os.MkdirAll(binDir, permDir); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	if err := os.MkdirAll(textDir, permDir); err != nil {
		return fmt.Errorf("failed to create text directory: %w", err)
	}

	// Create benchmark subdirectories
	for _, benchmark := range benchmarks {
		if err := os.MkdirAll(filepath.Join(binDir, benchmark), permDir); err != nil {
			return fmt.Errorf("failed to create bin subdirectory for %s: %w", benchmark, err)
		}
		if err := os.MkdirAll(filepath.Join(textDir, benchmark), permDir); err != nil {
			return fmt.Errorf("failed to create text subdirectory for %s: %w", benchmark, err)
		}
	}

	// Create description file
	if err := os.WriteFile(descFile, []byte(""), permFile); err != nil {
		return fmt.Errorf("failed to create description file: %w", err)
	}

	log.Printf("Created directory structure: %s\n", tagDir)
	return nil
}

func createProfileFunctionDirectories(tag string, profiles, benchmarks []string) error {
	tagDir := filepath.Join("bench", tag)

	// Only create directories for pprof profiles (not trace)
	for _, profile := range profiles {
		if profile == "trace" {
			continue
		}

		profileDir := filepath.Join(tagDir, profile+"_functions")
		if err := os.MkdirAll(profileDir, permDir); err != nil {
			return fmt.Errorf("failed to create profile directory %s: %w", profileDir, err)
		}

		for _, benchmark := range benchmarks {
			benchmarkDir := filepath.Join(profileDir, benchmark)
			if err := os.MkdirAll(benchmarkDir, permDir); err != nil {
				return fmt.Errorf("failed to create benchmark directory %s: %w", benchmarkDir, err)
			}
		}
	}

	log.Printf("Created profile function directories\n")
	return nil
}

func RunBenchmark(benchmarkName string, profiles []string, count int, tag string) error {
	cmd := buildBenchmarkCommand(benchmarkName, profiles, count)
	textDir, binDir := getOutputDirectories(benchmarkName, tag)

	outputFile := filepath.Join(textDir, benchmarkName+".txt")
	if err := runBenchmarkCommand(cmd, outputFile); err != nil {
		return err
	}

	if err := moveProfileFiles(benchmarkName, profiles, binDir); err != nil {
		return err
	}

	return moveTestFiles(benchmarkName, binDir)
}

func buildBenchmarkCommand(benchmarkName string, profiles []string, count int) []string {
	cmd := []string{
		"go", "test", "-run=^$",
		fmt.Sprintf("-bench=^%s$", benchmarkName),
		"-benchmem",
		fmt.Sprintf("-count=%d", count),
	}

	for _, profile := range profiles {
		if flag, exists := profileFlags[profile]; exists {
			cmd = append(cmd, flag)
		}
	}

	return cmd
}

func getOutputDirectories(benchmarkName, tag string) (string, string) {
	tagDir := filepath.Join("bench", tag)
	textDir := filepath.Join(tagDir, "text", benchmarkName)
	binDir := filepath.Join(tagDir, "bin", benchmarkName)

	// Only create if not exists
	_ = os.MkdirAll(textDir, permDir)
	_ = os.MkdirAll(binDir, permDir)

	return textDir, binDir
}

func runBenchmarkCommand(cmd []string, outputFile string) error {
	execCmd := exec.Command(cmd[0], cmd[1:]...)

	output, err := execCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "go: cannot find main module") {
			return fmt.Errorf("go: cannot find main module - ensure you're in a Go project directory")
		}
		return fmt.Errorf("benchmark failed: %s", string(output))
	}

	return os.WriteFile(outputFile, output, permFile)
}

func moveProfileFiles(benchmarkName string, profiles []string, binDir string) error {
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

		// Wait for file to be fully written (still a workaround)
		time.Sleep(100 * time.Millisecond)

		newPath := filepath.Join(binDir, fmt.Sprintf("%s_%s.out", benchmarkName, profile))
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
		if err := os.Rename(file, newPath); err != nil {
			return fmt.Errorf("failed to move test file %s: %w", file, err)
		}
	}
	return nil
}

func ProcessProfiles(benchmarkName string, profiles []string, tag string) error {
	tagDir := filepath.Join("bench", tag)
	binDir := filepath.Join(tagDir, "bin", benchmarkName)
	textDir := filepath.Join(tagDir, "text", benchmarkName)

	for _, profile := range profiles {
		if profile == "trace" {
			continue
		}

		profileFile := filepath.Join(binDir, fmt.Sprintf("%s_%s.out", benchmarkName, profile))
		if _, err := os.Stat(profileFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Printf("Warning: Profile file not found: %s\n", profileFile)
				continue
			}
			return fmt.Errorf("failed to stat profile file %s: %w", profileFile, err)
		}

		outputFile := filepath.Join(textDir, fmt.Sprintf("%s_%s.txt", benchmarkName, profile))
		profileFunctionsDir := filepath.Join(tagDir, profile+"_functions", benchmarkName)

		if err := generateTextProfile(profileFile, outputFile); err != nil {
			return fmt.Errorf("failed to generate text profile for %s: %w", profile, err)
		}

		pngFile := filepath.Join(profileFunctionsDir, fmt.Sprintf("%s_%s.png", benchmarkName, profile))
		if err := generatePNGVisualization(profileFile, pngFile); err != nil {
			return fmt.Errorf("failed to generate PNG visualization for %s: %w", profile, err)
		}

		log.Printf("Processed %s profile for %s\n", profile, benchmarkName)
	}

	return nil
}

func generateTextProfile(profileFile, outputFile string) error {
	cmd := append([]string{"go", "tool", "pprof"}, pprofTextParams...)
	cmd = append(cmd, profileFile)

	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof command failed: %w", err)
	}

	return os.WriteFile(outputFile, output, permFile)
}

func generatePNGVisualization(profileFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", "-png", profileFile}

	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof PNG generation failed: %w", err)
	}

	return os.WriteFile(outputFile, output, permFile)
}

func AnalyzeProfileFunctions(tag string, profiles []string, benchmarkName string, benchmarkConfig config.BenchmarkFilter) error {
	for _, profile := range profiles {
		if profile == "trace" {
			continue
		}

		paths := getProfilePaths(tag, benchmarkName, profile)
		if err := os.MkdirAll(paths.FunctionDirectory, permDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		filter := parser.ProfileFilter{
			FunctionPrefixes: benchmarkConfig.Prefixes,
			IgnoreFunctions:  parseIgnoreList(benchmarkConfig.Ignore),
		}

		// TODO: Does it really need to happen in two steps ?
		functions, err := parser.GetAllFunctionNames(paths.ProfileTextFile, filter)
		if err != nil {
			return fmt.Errorf("failed to extract function names: %w", err)
		}
		//
		for _, function := range functions {
			if err := getFunctionPprofContent(function, paths); err != nil {
				return fmt.Errorf("failed to extract function content for %s: %w", function, err)
			}
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
	return ProfilePaths{
		ProfileTextFile:   filepath.Join(tagDir, "text", benchmarkName, fmt.Sprintf("%s_%s.txt", benchmarkName, profile)),
		ProfileBinaryFile: filepath.Join(tagDir, "bin", benchmarkName, fmt.Sprintf("%s_%s.out", benchmarkName, profile)),
		FunctionDirectory: filepath.Join(tagDir, profile+"_functions", benchmarkName),
	}
}

// getFunctionPprofContent gets code line level mapping of specified function
// and writes the data to a file named after the function.
func getFunctionPprofContent(function string, paths ProfilePaths) error {
	outputFile := filepath.Join(paths.FunctionDirectory, function+".txt")
	cmd := []string{"go", "tool", "pprof", fmt.Sprintf("-list=%s", function), paths.ProfileBinaryFile}

	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof list command failed: %w", err)
	}

	if err := os.WriteFile(outputFile, output, permFile); err != nil {
		return fmt.Errorf("failed to write function content: %w", err)
	}

	log.Printf("Collected function %s\n", function)
	return nil
}

// parseIgnoreList receives a comma separated
// string and turning it into a string slice.
func parseIgnoreList(ignore string) []string {
	if ignore == "" {
		return nil
	}

	parts := strings.Split(ignore, ",")
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

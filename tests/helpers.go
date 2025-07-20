package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/shared"
)

var (
	envDirName = envDirNameStatic
)

type IsFileExpected bool
type fileFullName string

const (
	Expected   IsFileExpected = true
	Unexpected IsFileExpected = false
)

const (
	envDirNameStatic = "Enviroment"
	templateFile     = "config_template.json"
	testDirName      = "tests"
	tag              = "test_tag"
	count            = "1"
	cpuProfile       = "cpu"
	memProfile       = "memory"
	benchName        = "BenchmarkStringProcessor"
)

func createPackage(dir string) error {
	// Create utils package directory
	utilsDir := filepath.Join(dir, "utils")
	if err := os.MkdirAll(utilsDir, shared.PermDir); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	// Create the utils package
	packageContent := `package utils

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// StringProcessor provides string manipulation utilities
type StringProcessor struct {
	data []string
}

// NewStringProcessor creates a new string processor
func NewStringProcessor() *StringProcessor {
	return &StringProcessor{
		data: make([]string, 0),
	}
}

// AddString adds a string to the processor
func (sp *StringProcessor) AddString(s string) {
	sp.data = append(sp.data, s)
}

// ProcessStrings performs various operations on stored strings
func (sp *StringProcessor) ProcessStrings() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Sort strings
	sorted := make([]string, len(sp.data))
	copy(sorted, sp.data)
	sort.Strings(sorted)
	result["sorted"] = sorted
	
	// Calculate total length
	totalLen := 0
	for _, s := range sp.data {
		totalLen += len(s)
	}
	result["total_length"] = totalLen
	
	// Generate hashes
	hashes := make([]string, len(sp.data))
	for i, s := range sp.data {
		hash := sha256.Sum256([]byte(s))
		hashes[i] = fmt.Sprintf("%x", hash)
	}
	result["hashes"] = hashes
	
	return result
}

// Calculator provides mathematical operations
type Calculator struct{}

// NewCalculator creates a new calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Fibonacci calculates fibonacci number (CPU intensive)
func (c *Calculator) Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return c.Fibonacci(n-1) + c.Fibonacci(n-2)
}

// MatrixMultiply performs matrix multiplication
func (c *Calculator) MatrixMultiply(a, b [][]int) [][]int {
	if len(a) == 0 || len(b) == 0 || len(a[0]) != len(b) {
		return nil
	}
	
	rows, cols := len(a), len(b[0])
	result := make([][]int, rows)
	for i := range result {
		result[i] = make([]int, cols)
	}
	
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			for k := 0; k < len(b); k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	
	return result
}

// DataGenerator generates test data
type DataGenerator struct{}

// NewDataGenerator creates a new data generator
func NewDataGenerator() *DataGenerator {
	return &DataGenerator{}
}

// GenerateStrings creates a slice of strings
func (dg *DataGenerator) GenerateStrings(count int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = fmt.Sprintf("generated_string_%d_%s", i, strings.Repeat("x", i%100))
	}
	return result
}

// GenerateMatrix creates a matrix of given size
func (dg *DataGenerator) GenerateMatrix(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
		for j := range matrix[i] {
			matrix[i][j] = (i + j) % 100
		}
	}
	return matrix
}
`

	utilsPath := filepath.Join(utilsDir, "utils.go")
	return os.WriteFile(utilsPath, []byte(packageContent), shared.PermFile)
}

func createBenchmarkFile(dir string) error {
	benchmarkContent := `package main

import (
	"testing"
	"test-environment/utils"
)

func BenchmarkStringProcessor(b *testing.B) {
	processor := utils.NewStringProcessor()
	generator := utils.NewDataGenerator()
	
	// Generate test data
	strings := generator.GenerateStrings(1000)
	for _, s := range strings {
		processor.AddString(s)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := processor.ProcessStrings()
		_ = result
	}
}

func BenchmarkFibonacci(b *testing.B) {
	calc := utils.NewCalculator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := calc.Fibonacci(25)
		_ = result
	}
}

func BenchmarkMatrixMultiplication(b *testing.B) {
	calc := utils.NewCalculator()
	generator := utils.NewDataGenerator()
	
	// Generate test matrices
	matrixA := generator.GenerateMatrix(50, 50)
	matrixB := generator.GenerateMatrix(50, 50)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := calc.MatrixMultiply(matrixA, matrixB)
		_ = result
	}
}

func BenchmarkDataGeneration(b *testing.B) {
	generator := utils.NewDataGenerator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strings := generator.GenerateStrings(500)
		matrix := generator.GenerateMatrix(20, 20)
		_ = strings
		_ = matrix
	}
}
`

	benchPath := filepath.Join(dir, "benchmark_test.go")
	return os.WriteFile(benchPath, []byte(benchmarkContent), shared.PermFile)
}

func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func createConfigFile(t *testing.T, cfgTemplate *config.Config) {
	t.Helper()

	configPath := filepath.Join(envDirName, templateFile)

	data, err := json.MarshalIndent(cfgTemplate, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal config template: %v", err)
	}

	if err := os.WriteFile(configPath, data, shared.PermFile); err != nil {
		t.Fatalf("failed to write config template file: %v", err)
	}
}

func setupEnviroment(t *testing.T) {
	// 1. Create environment Directory.
	if err := os.Mkdir(envDirName, shared.PermDir); err != nil && !os.IsExist(err) {
		t.Fatalf("couldn't create environment dir: %v", err)
	}

	// 2. Initialize Go module.
	cmd := exec.Command("go", "mod", "init", "test-environment")
	cmd.Dir = envDirName

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to initialize Go module: %v\nOutput: %s", err, output)
	}

	// 3. Create package and benchmark files.
	if err := createPackage(envDirName); err != nil {
		t.Fatalf("failed to create package: %v", err)
	}

	if err := createBenchmarkFile(envDirName); err != nil {
		t.Fatalf("failed to create benchmark file: %v", err)
	}
}

func setUpProf(t *testing.T, projectRoot string) {
	t.Helper()

	// Build prof binary directly to the environment directory
	profBinary := filepath.Join(projectRoot, testDirName, envDirName, "prof")

	// Build from cmd/prof directory
	cmdProfDir := filepath.Join(projectRoot, "cmd", "prof")
	buildCmd := exec.Command("go", "build", "-o", profBinary, ".")
	buildCmd.Dir = cmdProfDir // Run from cmd/prof directory

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build prof binary: %v\nOutput: %s", err, buildOutput)
	}
}

func runProf(t *testing.T, projectRoot string, args []string) {
	t.Helper()

	envFullPath := filepath.Join(projectRoot, testDirName, envDirName)
	profBinary := filepath.Join(envFullPath, "prof")

	if _, err := os.Stat(profBinary); os.IsNotExist(err) {
		t.Fatalf("prof binary not found at: %s", profBinary)
	}

	cmd := exec.Command("./prof", args...)
	cmd.Dir = envFullPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("prof command failed: %v\nStdout: %s\nStderr: %s",
			err, stdout.String(), stderr.String())
	}

	successMessage := shared.InfoCollectionSuccess
	if !strings.Contains(stderr.String(), successMessage) {
		t.Fatalf("Expected success message not found")
	}
}

func checkDirectory(t *testing.T, path, description string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s does not exist: %s", description, path)
	}
}

func checkOutput(t *testing.T, envPath string, profiles []string, expectNonSpecifiedFiles, withConfig bool, expectedFiles map[fileFullName]IsFileExpected) {
	t.Helper()

	benchPath := filepath.Join(envPath, shared.MainDirOutput)

	// 1. Check that the tag dir exists
	tagPath := filepath.Join(benchPath, tag)
	checkDirectory(t, tagPath, tag)

	// 2. Check that the bin and text dir exists
	binPath := filepath.Join(tagPath, shared.ProfileBinDir)
	textPath := filepath.Join(tagPath, shared.ProfileTextDir)
	binBenchPath := filepath.Join(binPath, benchName)
	textBenchPath := filepath.Join(textPath, benchName)

	// bin => cpu, mem, test
	// txt => cpu, mem, bench
	expectedNumberOfFiles := 3

	var doesConfigApply bool

	checkDirectory(t, binPath, "bin directory")
	checkDirectory(t, binBenchPath, "benchmark directory inside of bin")
	checkDirectoryFiles(t, binBenchPath, "bin files inside of benchmark directory", expectedNumberOfFiles, expectNonSpecifiedFiles, doesConfigApply, expectedFiles)

	checkDirectory(t, textPath, "text directory")
	checkDirectory(t, textBenchPath, "benchmark directory inside of text")
	checkDirectoryFiles(t, textBenchPath, "text files inside of benchmark directory", expectedNumberOfFiles, expectNonSpecifiedFiles, doesConfigApply, expectedFiles)

	// 3. Check that profile function directories and files exist for each profile
	for _, profile := range profiles {
		profileFunctionsDir := fmt.Sprintf("%s%s", profile, shared.FunctionsDirSuffix)
		profileFunctionsPath := filepath.Join(tagPath, profileFunctionsDir)
		checkDirectory(t, profileFunctionsPath, "profile functions directory e.g cpu_functions")

		benchDir := filepath.Join(profileFunctionsPath, benchName)
		checkDirectory(t, benchDir, "benchmark directory inside profile functions directory e.g cpu_functions/BenchmarkName")

		notSure := 0
		doesConfigApply = withConfig
		checkDirectoryFiles(t, benchDir, "individual function files inside benchmark directory", notSure, expectNonSpecifiedFiles, withConfig, expectedFiles)
	}

	if withConfig {
		remainingFiles := countExpectedFiles(expectedFiles)
		allowedRemainingFiles := 1

		if remainingFiles > allowedRemainingFiles {
			t.Fatalf("Expected almost all files to be found, %d files remaining", remainingFiles)
		}
	}
}

func countExpectedFiles(files map[fileFullName]IsFileExpected) int {
	count := 0
	for _, isExpected := range files {
		if bool(isExpected) {
			count++
		}
	}
	return count
}

func getFileOrDirs(t *testing.T, dirPath, dirDescription string) ([]os.DirEntry, []string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", dirDescription, err)
	}

	var files []os.DirEntry
	var dirNames []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		} else {
			files = append(files, entry)
		}
	}

	return files, dirNames
}

func checkDirectoryFiles(t *testing.T, dirPath, dirDescription string, expectedFileNum int, expectNonSpecifiedFiles, withConfig bool, specifiedFiles map[fileFullName]IsFileExpected) {
	t.Helper()

	files, dirNames := getFileOrDirs(t, dirPath, dirDescription)

	// Ensure no directories exist
	foundDirectories := len(dirNames)
	if foundDirectories > 0 {
		t.Fatalf("%s contains unexpected directories: %v", dirDescription, dirNames)
	}

	// Ensure files exist
	foundFiles := len(files)
	if foundFiles == 0 {
		t.Fatalf("%s contains no files", dirDescription)
	}

	if expectedFileNum > 0 && foundFiles != expectedFileNum {
		t.Fatalf("expected %d, found %d files", expectedFileNum, foundFiles)
	}

	// across all profile, all expected files must exist
	for _, file := range files {
		fileName := file.Name()
		if withConfig {
			fileKey := fileFullName(fileName)
			isExpected, ok := specifiedFiles[fileKey]
			if ok {
				if bool(isExpected) {
					delete(specifiedFiles, fileKey)
					continue
				} else {
					if !expectNonSpecifiedFiles {
						t.Fatalf("File should not be here: %s", fileKey)
					}
				}
			}

		}
		filePath := filepath.Join(dirPath, fileName)
		checkFileNotEmpty(t, filePath, fileName)
	}
}

func checkFileNotEmpty(t *testing.T, filePath, fileName string) {
	t.Helper()

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", fileName, err)
	}

	if fileInfo.Size() == 0 {
		t.Fatalf("File %s is empty (0 bytes)", fileName)
	}
}

func testConfigScenario(t *testing.T, cfg *config.Config, expectNonSpecifiedFiles, withConfig, withCleanUp bool, label string, specifiedFiles map[fileFullName]IsFileExpected) {
	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}

	envDirName = envDirName + " " + label
	envPath := path.Join(root, testDirName, envDirName)

	if withCleanUp {
		t.Cleanup(func() {
			envDirName = envDirNameStatic
			if err := os.RemoveAll(envPath); err != nil {
				t.Logf("Failed to clean up environment: %v", err)
			}
		})
	}

	// 1. Set up Environment
	setupEnviroment(t)

	// 2. Build prof and move to Environment
	createConfigFile(t, cfg)
	setUpProf(t, root)

	// 3. Run ./prof inside the Environment
	args := []string{
		"--benchmarks", fmt.Sprintf("[%s]", benchName),
		"--profiles", fmt.Sprintf("[%s,%s]", cpuProfile, memProfile),
		"--count", count,
		"--tag", tag,
	}

	runProf(t, root, args)

	// 4. Check bench output
	expectedProfiles := []string{cpuProfile, memProfile}
	checkOutput(t, envPath, expectedProfiles, expectNonSpecifiedFiles, withConfig, specifiedFiles)
}

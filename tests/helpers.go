package tests

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
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

type TestArgs struct {
	specifiedFiles          map[fileFullName]*FieldsCheck
	cfg                     config.Config
	withConfig              bool
	expectNonSpecifiedFiles bool
	noConfigFile            bool
	cmd                     []string
	expectedErrorMessage    string
	label                   string
	expectedNumberOfFiles   int
	withCleanUp             bool
	expectedProfiles        []string
}

type fileFullName string

type FieldsCheck struct {
	isFileExpected     bool
	minimumAppearances int
	maximumAppearances int
	currentAppearances int
}

func (fc *FieldsCheck) IsWithinRange() bool {
	return fc.currentAppearances >= fc.minimumAppearances &&
		fc.currentAppearances <= fc.maximumAppearances
}

func newDefaultFieldsCheckExpected() *FieldsCheck {
	return &FieldsCheck{
		isFileExpected:     true,
		minimumAppearances: 1,
		maximumAppearances: 2,
	}
}

func newDefaultFieldsCheckNotExpected() *FieldsCheck {
	return &FieldsCheck{
		isFileExpected: false,
	}
}

const (
	envDirNameStatic = "Enviroment"
	templateFile     = "config_template.json"
	testDirName      = "tests"
	tag              = "test_tag"
	count            = "5"
	cpuProfile       = "cpu"
	memProfile       = "memory"
	blockProfile     = "block"
	benchName        = "BenchmarkStringProcessor"
)

//go:embed assets/utils.go.txt
var utilsTemplate string

func createPackage(dir string) error {
	utilsDir := filepath.Join(dir, "utils")
	if err := os.MkdirAll(utilsDir, shared.PermDir); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	utilsPath := filepath.Join(utilsDir, "utils.go")
	return os.WriteFile(utilsPath, []byte(utilsTemplate), shared.PermFile)
}

//go:embed assets/benchmark_test.go.txt
var BenchmarkContent string

func createBenchmarkFile(dir string) error {
	benchPath := filepath.Join(dir, "benchmark_test.go")
	return os.WriteFile(benchPath, []byte(BenchmarkContent), shared.PermFile)
}

func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
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

	if err = os.WriteFile(configPath, data, shared.PermFile); err != nil {
		t.Fatalf("failed to write config template file: %v", err)
	}
}

func setupEnviroment(t *testing.T) {
	t.Helper()
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
	if err = createPackage(envDirName); err != nil {
		t.Fatalf("failed to create package: %v", err)
	}

	if err = createBenchmarkFile(envDirName); err != nil {
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

func runProf(t *testing.T, projectRoot string, args []string, expectedErrMessage string) (shouldContinue bool) {
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
		shouldContinue = handleCommandError(t, err, &stdout, &stderr, expectedErrMessage)
		if !shouldContinue {
			return false
		}
	}

	successMessage := shared.InfoCollectionSuccess
	if !strings.Contains(stderr.String(), successMessage) {
		t.Fatal("Expected success message not found")
	}

	return true
}

func handleCommandError(t *testing.T, err error, stdout, stderr *bytes.Buffer, expectedErrMessage string) bool {
	t.Helper()

	if expectedErrMessage != "" {
		stderrText := stderr.String()
		if strings.Contains(stderrText, expectedErrMessage) {
			return false // Expected error found, caller should return
		}

		t.Fatalf("Expected error message '%s' not found.\nStderr: %s\nStdout: %s",
			expectedErrMessage, stderrText, stdout.String())
	}

	t.Fatalf("prof command failed: %v\nStdout: %s\nStderr: %s",
		err, stdout.String(), stderr.String())

	return true // Never reached in case of t.Fatalf
}

func checkDirectory(t *testing.T, path, description string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s does not exist: %s", description, path)
	}
}

func checkOutput(t *testing.T, envPath string, testArgs *TestArgs) {
	t.Helper()

	expectedProfiles := testArgs.expectedProfiles
	expectNonSpecifiedFiles := testArgs.expectNonSpecifiedFiles
	withConfig := testArgs.withConfig
	specifiedFiles := testArgs.specifiedFiles
	expectedNumberOfFiles := testArgs.expectedNumberOfFiles

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
	configDoesntApply := false

	checkDirectory(t, binPath, "bin directory")
	checkDirectory(t, binBenchPath, "benchmark directory inside of bin")
	checkDirectoryFiles(t, binBenchPath, "bin files inside of benchmark directory", expectedNumberOfFiles, expectNonSpecifiedFiles, configDoesntApply, specifiedFiles)

	checkDirectory(t, textPath, "text directory")
	checkDirectory(t, textBenchPath, "benchmark directory inside of text")
	checkDirectoryFiles(t, textBenchPath, "text files inside of benchmark directory", expectedNumberOfFiles, expectNonSpecifiedFiles, configDoesntApply, specifiedFiles)

	// 3. Check that profile function directories and files exist for each profile
	for _, profile := range expectedProfiles {
		profileFunctionsDir := fmt.Sprintf("%s%s", profile, shared.FunctionsDirSuffix)
		profileFunctionsPath := filepath.Join(tagPath, profileFunctionsDir)
		checkDirectory(t, profileFunctionsPath, "profile functions directory e.g cpu_functions")

		benchDir := filepath.Join(profileFunctionsPath, benchName)
		checkDirectory(t, benchDir, "benchmark directory inside profile functions directory e.g cpu_functions/BenchmarkName")

		notSure := 0
		checkDirectoryFiles(t, benchDir, "individual function files inside benchmark directory", notSure, expectNonSpecifiedFiles, withConfig, specifiedFiles)
	}

	if withConfig {
		remainingFiles := countRemainingFiles(specifiedFiles)
		maxAllowRemainingFiles := 1

		if remainingFiles > maxAllowRemainingFiles {
			t.Fatalf("Expected almost all files to be found, %d files remaining", remainingFiles)
		}
	}
}

func countRemainingFiles(files map[fileFullName]*FieldsCheck) int {
	count := 0
	for _, fieldsCheck := range files {
		if fieldsCheck.isFileExpected {
			if !fieldsCheck.IsWithinRange() {
				count++
			}
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

func checkDirectoryFiles(t *testing.T, dirPath, dirDescription string, expectedFileNum int, expectNonSpecifiedFiles, withConfig bool, specifiedFiles map[fileFullName]*FieldsCheck) {
	t.Helper()

	files, dirNames := getFileOrDirs(t, dirPath, dirDescription)

	// Ensure no directories exist
	containsDirectories := len(dirNames) > 0
	if containsDirectories {
		t.Fatalf("%s contains unexpected directories: %v", dirDescription, dirNames)
	}

	// Ensure files exist
	missingFiles := len(files) == 0
	if missingFiles {
		t.Fatalf("%s contains no files", dirDescription)
	}

	// Check expected file count
	expectingSpecificNumber := expectedFileNum > 0
	differentThanExpected := len(files) != expectedFileNum
	if expectingSpecificNumber && differentThanExpected {
		t.Fatalf("expected %d, found %d files", expectedFileNum, len(files))
	}

	// Check each file
	for _, file := range files {
		fileName := file.Name()

		if withConfig {
			validateFileWithConfig(t, fileName, expectNonSpecifiedFiles, specifiedFiles)
		}

		filePath := filepath.Join(dirPath, fileName)
		checkFileNotEmpty(t, filePath, fileName)
	}
}

func validateFileWithConfig(t *testing.T, fileName string, expectNonSpecifiedFiles bool, specifiedFiles map[fileFullName]*FieldsCheck) {
	fileKey := fileFullName(fileName)
	fieldsCheck, exists := specifiedFiles[fileKey]

	if exists {
		if fieldsCheck.isFileExpected {
			fieldsCheck.currentAppearances++
		} else {
			t.Fatalf("File should have been filtered: %s", fileKey)
		}
	} else if !expectNonSpecifiedFiles {
		t.Fatalf("Unexpected file: %s", fileKey)
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

func testConfigScenario(t *testing.T, testArgs *TestArgs) {
	defer func() {
		envDirName = envDirNameStatic
	}()

	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}

	envDirName = envDirName + " " + testArgs.label
	envPath := path.Join(root, testDirName, envDirName)

	if testArgs.withCleanUp {
		t.Cleanup(func() {
			if err = os.RemoveAll(envPath); err != nil {
				t.Logf("Failed to clean up environment: %v", err)
			}
		})
	}

	// 1. Set up Environment
	setupEnviroment(t)

	// 2. Build prof and move to Environment
	if !testArgs.noConfigFile {
		createConfigFile(t, &testArgs.cfg)
	}

	setUpProf(t, root)

	shouldContinue := runProf(t, root, testArgs.cmd, testArgs.expectedErrorMessage)
	// Tested failure
	if !shouldContinue {
		return
	}

	// 4. Check bench output
	checkOutput(t, envPath, testArgs)
}

func defaultRunCmd() []string {
	return []string{
		"run",
		"--benchmarks", benchName,
		"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
		"--count", count,
		"--tag", tag,
	}
}

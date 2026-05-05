package tests

import (
	"bytes"
	_ "embed" // embedded assets for synthetic test modules (see //go:embed below)
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

// integrationEnvDir is the per-scenario directory name under tests/ (package cwd).
func integrationEnvDir(label string) string {
	return envDirNameStatic + " " + label
}

// TestArgs holds inputs and expectations for a single integration test scenario.
type TestArgs struct {
	specifiedFiles          map[fileFullName]*FieldsCheck
	cfg                     internal.Config
	expectedNumberOfFiles   int
	expectedErrorMessage    string
	label                   string
	cmd                     []string
	expectedProfiles        []string
	withConfig              bool
	expectNonSpecifiedFiles bool
	noConfigFile            bool
	withCleanUp             bool
	blockOutputCheck        bool
	isEnvironmentSet        bool
	checkSuccessMessage     bool
}

type fileFullName string

// FieldsCheck records expected min/max appearances for an output file in a test run.
type FieldsCheck struct {
	isFileExpected     bool
	minimumAppearances int
	maximumAppearances int
	currentAppearances int
}

// IsWithinRange reports whether currentAppearances is between minimum and maximum inclusive.
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
	// Multiple bench counts keep CPU sampling + filtering stable on CI (esp. WithFunctionFilterPlusIgnore).
	count        = "10"
	cpuProfile   = "cpu"
	memProfile   = "memory"
	blockProfile = "block"
	benchName    = "BenchmarkStringProcessor"
)

// profBinaryName is the built prof executable filename. On Windows, os/exec
// requires a recognized extension (e.g. .exe); a bare "prof" fails lookup.
func profBinaryName() string {
	if runtime.GOOS == "windows" {
		return "prof.exe"
	}
	return "prof"
}

//go:embed assets/utils.go.txt
var utilsTemplate string

func createPackage(dir string) error {
	utilsDir := filepath.Join(dir, "utils")
	if err := os.MkdirAll(utilsDir, internal.PermDir); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	utilsPath := filepath.Join(utilsDir, "utils.go")
	return os.WriteFile(utilsPath, []byte(utilsTemplate), internal.PermFile)
}

// BenchmarkContent is embedded benchmark test source used when creating synthetic modules.
//
//go:embed assets/benchmark_test.go.txt
var BenchmarkContent string

func createBenchmarkFile(dir string) error {
	benchPath := filepath.Join(dir, "benchmark_test.go")
	return os.WriteFile(benchPath, []byte(BenchmarkContent), internal.PermFile)
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

func createConfigFile(t *testing.T, envDir string, cfgTemplate *internal.Config) {
	t.Helper()

	configPath := filepath.Join(envDir, templateFile)

	data, err := json.MarshalIndent(cfgTemplate, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal config template: %v", err)
	}

	if err = os.WriteFile(configPath, data, internal.PermFile); err != nil {
		t.Fatalf("failed to write config template file: %v", err)
	}
}

func setupEnviroment(t *testing.T, envDir string) {
	t.Helper()
	// 1. Create environment Directory.
	if err := os.Mkdir(envDir, internal.PermDir); err != nil && !os.IsExist(err) {
		t.Fatalf("couldn't create environment dir: %v", err)
	}

	// 2. Initialize Go module (skip if go.mod already exists — e.g. committed fixture under tests/).
	goModPath := filepath.Join(envDir, "go.mod")
	if _, statErr := os.Stat(goModPath); statErr != nil {
		if !os.IsNotExist(statErr) {
			t.Fatalf("stat go.mod: %v", statErr)
		}
		cmd := exec.Command("go", "mod", "init", "test-environment")
		cmd.Dir = envDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to initialize Go module: %v\nOutput: %s", err, output)
		}
	}

	// 3. Create package and benchmark files.
	if err := createPackage(envDir); err != nil {
		t.Fatalf("failed to create package: %v", err)
	}

	if err := createBenchmarkFile(envDir); err != nil {
		t.Fatalf("failed to create benchmark file: %v", err)
	}
}

var (
	integrationProfOnce     sync.Once
	integrationProfCached   string
	errIntegrationProfBuild error
)

// profCacheDir is a single build output reused for all integration scenarios in one `go test` run.
const profCacheDir = ".integration_prof_build"

func copyProfBinary(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if st, statErr := os.Stat(src); statErr == nil {
			_ = os.Chmod(dst, st.Mode()&0o777)
		}
	}
	return nil
}

// ensureCachedProfBinary builds cmd/prof once per test process; callers copy into each env dir.
func ensureCachedProfBinary(projectRoot string) (string, error) {
	integrationProfOnce.Do(func() {
		cacheDir := filepath.Join(projectRoot, testDirName, profCacheDir)
		if err := os.MkdirAll(cacheDir, internal.PermDir); err != nil {
			errIntegrationProfBuild = err
			return
		}
		out := filepath.Join(cacheDir, profBinaryName())
		cmdProfDir := filepath.Join(projectRoot, "cmd", "prof")
		buildCmd := exec.Command("go", "build", "-o", out, ".")
		buildCmd.Dir = cmdProfDir
		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			errIntegrationProfBuild = fmt.Errorf("failed to build prof binary: %w\nOutput: %s", err, buildOutput)
			return
		}
		integrationProfCached = out
	})
	return integrationProfCached, errIntegrationProfBuild
}

func setUpProf(t *testing.T, projectRoot, envDir string) {
	t.Helper()

	dst := filepath.Join(projectRoot, testDirName, envDir, profBinaryName())
	src, err := ensureCachedProfBinary(projectRoot)
	if err != nil {
		t.Fatalf("prof cache build: %v", err)
	}
	if err = copyProfBinary(src, dst); err != nil {
		t.Fatalf("failed to copy prof binary: %v", err)
	}
}

func newProfCmd(t *testing.T, envFullPath string, args []string) *exec.Cmd {
	t.Helper()

	profBinary := filepath.Join(envFullPath, profBinaryName())
	if _, err := os.Stat(profBinary); os.IsNotExist(err) {
		t.Fatalf("prof binary not found at: %s", profBinary)
	}

	cmd := exec.Command(profBinary, args...)
	cmd.Dir = envFullPath
	return cmd
}

// runProfCaptured runs prof like runProf but returns stdout/stderr for assertions.
func runProfCaptured(t *testing.T, envFullPath string, args []string, expectedErrMessage string, checkSuccessMessage bool) (stdout, stderr string, shouldContinue bool) {
	t.Helper()

	var stdoutB, stderrB bytes.Buffer
	cmd := newProfCmd(t, envFullPath, args)
	cmd.Stdout = &stdoutB
	cmd.Stderr = &stderrB

	err := cmd.Run()
	if err != nil {
		shouldContinue = handleCommandError(t, err, &stdoutB, &stderrB, expectedErrMessage)
		return stdoutB.String(), stderrB.String(), shouldContinue
	}

	successMessage := internal.InfoCollectionSuccess
	if checkSuccessMessage && !strings.Contains(stderrB.String(), successMessage) {
		t.Fatal("Expected success message not found")
	}

	return stdoutB.String(), stderrB.String(), true
}

func runProf(t *testing.T, envFullPath string, args []string, expectedErrMessage string, checkSuccessMessage bool) (shouldContinue bool) {
	t.Helper()
	_, _, ok := runProfCaptured(t, envFullPath, args, expectedErrMessage, checkSuccessMessage)
	return ok
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

	benchPath := filepath.Join(envPath, internal.MainDirOutput)

	// 1. Check that the tag dir exists
	tagPath := filepath.Join(benchPath, tag)
	checkDirectory(t, tagPath, tag)

	// 2. Check that the bin and text dir exists
	binPath := filepath.Join(tagPath, internal.ProfileBinDir)
	textPath := filepath.Join(tagPath, internal.ProfileTextDir)
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
		profileFunctionsDir := fmt.Sprintf("%s%s", profile, internal.FunctionsDirSuffix)
		profileFunctionsPath := filepath.Join(tagPath, profileFunctionsDir)
		checkDirectory(t, profileFunctionsPath, "profile functions directory e.g cpu_functions")

		benchDir := filepath.Join(profileFunctionsPath, benchName)
		checkDirectory(t, benchDir, "benchmark directory inside profile functions directory e.g cpu_functions/BenchmarkName")

		notSure := 0
		checkDirectoryFiles(t, benchDir, "individual function files inside benchmark directory", notSure, expectNonSpecifiedFiles, withConfig, specifiedFiles)
	}
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
	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}

	envDir := integrationEnvDir(testArgs.label)
	envFullPath := filepath.Join(root, testDirName, envDir)

	if testArgs.withCleanUp {
		t.Cleanup(func() {
			if err = os.RemoveAll(envFullPath); err != nil {
				t.Logf("Failed to clean up environment: %v", err)
			}
		})
	}

	if !testArgs.isEnvironmentSet {
		// 1. Set up Environment
		setupEnviroment(t, envDir)

		// 2. Build prof and move to Environment
		if !testArgs.noConfigFile {
			createConfigFile(t, envDir, &testArgs.cfg)
		}

		setUpProf(t, root, envDir)
	}

	shouldContinue := runProf(t, envFullPath, testArgs.cmd, testArgs.expectedErrorMessage, testArgs.checkSuccessMessage)
	// Tested failure
	if !shouldContinue {
		return
	}

	// 4. Check bench output
	if !testArgs.blockOutputCheck {
		checkOutput(t, envFullPath, testArgs)
	}
}

func defaultRunCmd() []string {
	cmd := []string{
		internal.AUTOCMD,
		"--benchmarks", benchName,
		"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
		"--count", count,
		"--tag", tag,
	}
	return append(cmd, autoBenchSkipPNGArgs()...)
}

// autoBenchSkipPNGArgs avoids requiring Graphviz during integration tests while `prof auto`
// defaults to strict PNG generation.
func autoBenchSkipPNGArgs() []string {
	return []string{"--skip-png"}
}

func buildProf(t *testing.T, outputPath, root string) {
	t.Helper()
	src, err := ensureCachedProfBinary(root)
	if err != nil {
		t.Fatalf("prof cache build: %v", err)
	}
	if err = copyProfBinary(src, outputPath); err != nil {
		t.Fatalf("failed to copy prof binary: %v", err)
	}
}

func createBenchForTracker(t *testing.T, label, iterations, tagName string, blockOutputCheck, isEnvironmentSet bool) {
	cmd := []string{
		internal.AUTOCMD,
		"--benchmarks", benchName,
		"--profiles", "cpu",
		"--count", iterations,
		"--tag", tagName,
	}
	cmd = append(cmd, autoBenchSkipPNGArgs()...)

	testArgs := &TestArgs{
		specifiedFiles:          nil,
		cfg:                     internal.Config{},
		withConfig:              false,
		expectNonSpecifiedFiles: true,
		noConfigFile:            true,
		cmd:                     cmd,
		expectedErrorMessage:    "",
		label:                   label,
		expectedNumberOfFiles:   3,
		withCleanUp:             false,
		expectedProfiles:        nil,
		blockOutputCheck:        blockOutputCheck,
		isEnvironmentSet:        isEnvironmentSet,
	}

	testConfigScenario(t, testArgs)
}

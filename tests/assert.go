package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

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

// functionFile builds the per-function output filename used as a key in
// specifiedFiles maps (e.g. "ProcessStrings.txt").
func functionFile(name string) fileFullName {
	return fileFullName(name + "." + workspace.TextExtension)
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

	benchPath := filepath.Join(envPath, workspace.MainDirOutput)

	tagPath := filepath.Join(benchPath, tag)
	checkDirectory(t, tagPath, tag)

	binPath := filepath.Join(tagPath, workspace.ProfileBinDir)
	textPath := filepath.Join(tagPath, workspace.ProfileTextDir)
	binBenchPath := filepath.Join(binPath, benchName)
	textBenchPath := filepath.Join(textPath, benchName)

	configDoesntApply := false

	checkDirectory(t, binPath, "bin directory")
	checkDirectory(t, binBenchPath, "benchmark directory inside of bin")
	checkDirectoryFiles(t, binBenchPath, "bin files inside of benchmark directory", expectedNumberOfFiles, expectNonSpecifiedFiles, configDoesntApply, specifiedFiles)

	checkDirectory(t, textPath, "text directory")
	checkDirectory(t, textBenchPath, "benchmark directory inside of text")
	checkDirectoryFiles(t, textBenchPath, "text files inside of benchmark directory", expectedNumberOfFiles, expectNonSpecifiedFiles, configDoesntApply, specifiedFiles)

	for _, profile := range expectedProfiles {
		profileFunctionsDir := fmt.Sprintf("%s%s", profile, workspace.FunctionsDirSuffix)
		profileFunctionsPath := filepath.Join(tagPath, profileFunctionsDir)
		checkDirectory(t, profileFunctionsPath, "profile functions directory e.g cpu_functions")

		benchDir := filepath.Join(profileFunctionsPath, benchName)
		checkDirectory(t, benchDir, "benchmark directory inside profile functions directory e.g cpu_functions/BenchmarkName")

		// Per-function dir size depends on filter sampling jitter so we don't enforce a count.
		checkDirectoryFiles(t, benchDir, "individual function files inside benchmark directory", 0, expectNonSpecifiedFiles, withConfig, specifiedFiles)
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

	if len(dirNames) > 0 {
		t.Fatalf("%s contains unexpected directories: %v", dirDescription, dirNames)
	}

	if len(files) == 0 {
		t.Fatalf("%s contains no files", dirDescription)
	}

	if expectedFileNum > 0 && len(files) != expectedFileNum {
		t.Fatalf("expected %d, found %d files", expectedFileNum, len(files))
	}

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

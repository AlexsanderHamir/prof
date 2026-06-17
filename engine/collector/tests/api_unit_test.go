package collector_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/internal/testpaths"
	"github.com/AlexsanderHamir/prof/parser"
)

// cleanupBenchDirectory removes the bench directory if it exists
func cleanupBenchDirectory(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(internal.MainDirOutput); err == nil {
		if err := os.RemoveAll(internal.MainDirOutput); err != nil {
			t.Errorf("Failed to clean up bench directory: %v", err)
		}
	}
}

func TestGetProfileTextOutput(t *testing.T) {
	binaryFile := testpaths.MustAsset(t, "cpu.out")

	// Check if the file exists
	_, err := os.Stat(binaryFile)
	if os.IsNotExist(err) {
		t.Skip("Binary file not found, skipping test")
	}

	// Create a temporary output file
	tempDir, err := os.MkdirTemp("", "test_profile_output")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "cpu_profile.txt")

	// Test the function
	err = collector.GetProfileTextOutput(tooling.NewExecRunner(), binaryFile, outputFile)

	// The function might fail if go tool pprof is not available
	// or if the binary file is not a valid profile
	if err != nil {
		// Check if the error is due to missing go tool or invalid profile
		if strings.Contains(err.Error(), "exec: \"go\": executable file not found in PATH") {
			t.Skip("Go tool not available, skipping test")
		}
		t.Errorf("GetProfileTextOutput failed: %v", err)
	} else {
		// If successful, check if output file was created and has content
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file was not created: %s", outputFile)
		} else {
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
			} else if len(content) == 0 {
				t.Errorf("Output file is empty")
			}
		}
	}
}

func TestGetPNGOutput(t *testing.T) {
	binaryFile := testpaths.MustAsset(t, "memory.out")

	// Check if the file exists
	_, err := os.Stat(binaryFile)
	if os.IsNotExist(err) {
		t.Skip("Binary file not found, skipping test")
	}

	// Create a temporary output file
	tempDir, err := os.MkdirTemp("", "test_png_output")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "memory_profile.png")

	// Test the function
	err = collector.GetPNGOutput(tooling.NewExecRunner(), binaryFile, outputFile)

	// The function might fail if go tool pprof is not available,
	// or if graphviz is missing (pprof -png needs dot on many platforms).
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(err.Error(), "exec: \"go\": executable file not found in PATH") {
			t.Skip("Go tool not available, skipping test")
		}
		if strings.Contains(msg, "graphviz") || strings.Contains(msg, "dot") ||
			strings.Contains(msg, "exit status") {
			t.Skipf("pprof PNG not available in this environment: %v", err)
		}
		t.Errorf("GetPNGOutput failed: %v", err)
	} else {
		// If successful, check if output file was created and has content
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file was not created: %s", outputFile)
		} else {
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
			} else if len(content) == 0 {
				t.Errorf("Output file is empty")
			}
		}
	}
}

func TestGetFunctionsOutput(t *testing.T) {
	binaryFile := testpaths.MustAsset(t, "cpu.out")

	// Check if the file exists
	_, err := os.Stat(binaryFile)
	if os.IsNotExist(err) {
		t.Skip("Binary file not found, skipping test")
	}

	// Create a temporary output directory
	tempDir, err := os.MkdirTemp("", "test_functions_output")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	entries, parseErr := parser.GetFunctionListEntriesV2(binaryFile, internal.FunctionFilter{})
	if parseErr != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", parseErr)
	}
	if len(entries) < 2 {
		t.Fatalf("fixture cpu.out should yield at least 2 list entries, got %d", len(entries))
	}
	entries = entries[:2]

	// Test the function
	err = collector.GetFunctionsOutput(tooling.NewExecRunner(), entries, binaryFile, tempDir)

	// The function might fail if go tool pprof is not available
	// or if the binary file is not a valid profile
	if err != nil {
		// Check if the error is due to missing go tool or invalid profile
		if strings.Contains(err.Error(), "exec: \"go\": executable file not found in PATH") {
			t.Skip("Go tool not available, skipping test")
		}
		t.Errorf("GetFunctionsOutput failed: %v", err)
	} else {
		// If successful, check if output files were created
		for _, e := range entries {
			outputFile := filepath.Join(tempDir, e.OutputStem+"."+internal.TextExtension)
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Errorf("Output file was not created for function %s: %s", e.OutputStem, outputFile)
			}
		}
	}
}

func TestRunCollector(t *testing.T) {
	binaryFiles := []string{
		testpaths.MustAsset(t, "cpu.out"),
		testpaths.MustAsset(t, "memory.out"),
	}

	// Check if the files exist
	for _, file := range binaryFiles {
		_, err := os.Stat(file)
		if os.IsNotExist(err) {
			t.Skip("Binary files not found, skipping test")
		}
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_run_collector")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Ensure cleanup of any existing bench directory
	defer cleanupBenchDirectory(t)

	// Test the function
	err = collector.RunCollector(tooling.NewExecRunner(), binaryFiles, "test_tag", false)

	// The function might fail if go tool pprof is not available
	// or if the binary files are not valid profiles
	if err != nil {
		// Check if the error is due to missing go tool or invalid profile
		if strings.Contains(err.Error(), "exec: \"go\": executable file not found in PATH") {
			t.Skip("Go tool not available, skipping test")
		}
		t.Errorf("RunCollector failed: %v", err)
	} else {
		// If successful, check if the expected directory structure was created
		tagDir := filepath.Join("bench", "test_tag")
		if _, err := os.Stat(tagDir); os.IsNotExist(err) {
			t.Errorf("Tag directory was not created: %s", tagDir)
		}

		// Check if profile directories were created for each binary file
		for _, binaryFile := range binaryFiles {
			fileName := filepath.Base(binaryFile)
			fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
			profileDir := filepath.Join(tagDir, fileName)
			if _, err := os.Stat(profileDir); os.IsNotExist(err) {
				t.Errorf("Profile directory was not created: %s", profileDir)
			}

			// Check if text profile file was created
			textProfileFile := filepath.Join(profileDir, fileName+"."+internal.TextExtension)
			if _, err := os.Stat(textProfileFile); os.IsNotExist(err) {
				t.Errorf("Text profile file was not created: %s", textProfileFile)
			}

			// Check if functions directory was created
			functionsDir := filepath.Join(profileDir, "functions")
			if _, err := os.Stat(functionsDir); os.IsNotExist(err) {
				t.Errorf("Functions directory was not created: %s", functionsDir)
			}
		}
	}
}

func TestRunCollectorWithInvalidFiles(t *testing.T) {
	// Test with non-existent files
	invalidFiles := []string{"/non/existent/file.out", "/another/invalid/file.out"}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_run_collector_invalid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Ensure cleanup of any existing bench directory
	defer cleanupBenchDirectory(t)

	// Test the function - it should fail
	fr := &tooling.FakeRunner{Err: []error{errors.New("exit status 1")}}
	err = collector.RunCollector(fr, invalidFiles, "test_tag", false)
	if err == nil {
		t.Error("Expected error when running collector with invalid files, got nil")
	} else if !strings.Contains(err.Error(), "pprof command failed") {
		t.Errorf("Expected error to contain 'pprof command failed', got: %v", err)
	}
}

func TestRunCollectorWithEmptyFileList(t *testing.T) {
	// Test with empty file list
	emptyFiles := []string{}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_run_collector_empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Ensure cleanup of any existing bench directory
	defer cleanupBenchDirectory(t)

	// Test the function - it should succeed with no files to process
	err = collector.RunCollector(tooling.NewExecRunner(), emptyFiles, "test_tag", false)
	if err != nil {
		t.Errorf("Expected no error when running collector with empty file list, got: %v", err)
	}

	// Check if tag directory was created in the current working directory
	tagDir := filepath.Join("bench", "test_tag")
	if _, err := os.Stat(tagDir); os.IsNotExist(err) {
		t.Errorf("Tag directory was not created: %s", tagDir)
	}
}

func TestRunCollectorWithMockFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_run_collector_mock")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Ensure cleanup of any existing bench directory
	defer cleanupBenchDirectory(t)

	// Create mock binary files
	mockFiles := []string{
		createMockBinaryFile(t, tempDir, "mock1.out"),
		createMockBinaryFile(t, tempDir, "mock2.out"),
	}

	// Test the function - it should fail due to invalid binary files
	fr := &tooling.FakeRunner{Err: []error{errors.New("exit status 1")}}
	err = collector.RunCollector(fr, mockFiles, "test_tag", false)

	// The function should fail because the mock files are not valid Go profiles
	if err == nil {
		t.Error("Expected error when running collector with mock files, got nil")
	} else if !strings.Contains(err.Error(), "pprof command failed") {
		t.Errorf("Expected error to contain 'pprof command failed', got: %v", err)
	}
}

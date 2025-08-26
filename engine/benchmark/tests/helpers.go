package tests

import (
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
)

func createTestGoModule(root string) error {
	// Create go.mod file
	goModContent := `module github.com/test/benchmark

go 1.21
`
	goModPath := filepath.Join(root, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return err
	}

	// Create a test file with benchmark functions
	testFile := filepath.Join(root, "benchmark_test.go")
	testContent := `package main

import "testing"

func BenchmarkStringProcessor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "test string"
	}
}

func BenchmarkNumberCruncher(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = i * 2
	}
}

func TestSomething(t *testing.T) {
	t.Log("This is a regular test")
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		return err
	}

	// Create a subdirectory with another test file
	subDir := filepath.Join(root, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		return err
	}

	subTestFile := filepath.Join(subDir, "sub_benchmark_test.go")
	subTestContent := `package subdir

import "testing"

func BenchmarkSubProcessor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "sub test"
	}
}
`
	if err := os.WriteFile(subTestFile, []byte(subTestContent), 0644); err != nil {
		return err
	}

	return nil
}

// cleanupBenchDirectories removes any bench directories created during testing
func cleanupBenchDirectories() {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	// Remove the entire bench directory if it exists in current directory
	benchDir := filepath.Join(currentDir, internal.MainDirOutput)
	if err := os.RemoveAll(benchDir); err != nil {
		// Ignore errors during cleanup - file might not exist
	}

	// Also try to clean up in the tests subdirectory if we're running from there
	testsBenchDir := filepath.Join(currentDir, "tests", internal.MainDirOutput)
	if err := os.RemoveAll(testsBenchDir); err != nil {
		// Ignore errors during cleanup - file might not exist
	}

	// Try to clean up in the benchmark package directory
	benchmarkBenchDir := filepath.Join(currentDir, "engine", "benchmark", "tests", internal.MainDirOutput)
	if err := os.RemoveAll(benchmarkBenchDir); err != nil {
		// Ignore errors during cleanup - file might not exist
	}
}

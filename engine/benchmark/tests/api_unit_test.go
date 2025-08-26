package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/benchmark"
)

func TestRunBenchmarks(t *testing.T) {
	tests := []struct {
		name       string
		benchmarks []string
		profiles   []string
		tag        string
		count      int
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "empty benchmarks should return error",
			benchmarks: []string{},
			profiles:   []string{"cpu"},
			tag:        "test",
			count:      5,
			wantErr:    true,
			errMsg:     "benchmarks flag is empty",
		},
		{
			name:       "empty profiles should return error",
			benchmarks: []string{"BenchmarkTest"},
			profiles:   []string{},
			tag:        "test",
			count:      5,
			wantErr:    true,
			errMsg:     "profiles flag is empty",
		},
		{
			name:       "valid parameters should return error for non-existent benchmark",
			benchmarks: []string{"BenchmarkTest"},
			profiles:   []string{"cpu", "memory"},
			tag:        "test",
			count:      5,
			wantErr:    true,
			errMsg:     "failed to locate benchmark BenchmarkTest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanupBenchDirectories()

			err := benchmark.RunBenchmarks(tt.benchmarks, tt.profiles, tt.tag, tt.count, false)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RunBenchmarks() expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("RunBenchmarks() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("RunBenchmarks() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestDiscoverBenchmarks(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test Go module structure
	if err := createTestGoModule(tempDir); err != nil {
		t.Fatalf("Failed to create test Go module: %v", err)
	}

	tests := []struct {
		name             string
		scope            string
		wantErr          bool
		expectBenchmarks bool
		expectedCount    int
	}{
		{
			name:             "discover benchmarks in specific scope",
			scope:            tempDir,
			wantErr:          false,
			expectBenchmarks: true,
			expectedCount:    3, // BenchmarkStringProcessor, BenchmarkNumberCruncher, BenchmarkSubProcessor
		},
		{
			name:             "discover benchmarks in empty scope (module root)",
			scope:            "",
			wantErr:          false,
			expectBenchmarks: false,
			expectedCount:    0, // No benchmarks in actual module root
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			benchmarks, err := benchmark.DiscoverBenchmarks(tt.scope)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DiscoverBenchmarks() expected error but got none")
					return
				}
			} else {
				if err != nil {
					t.Errorf("DiscoverBenchmarks() unexpected error = %v", err)
					return
				}

				if tt.expectBenchmarks {
					// We expect at least our test benchmarks to be found
					if len(benchmarks) < tt.expectedCount {
						t.Errorf("DiscoverBenchmarks() returned %d benchmarks, want at least %d", len(benchmarks), tt.expectedCount)
					}

					// Check that we found the expected benchmark names
					expectedNames := map[string]bool{
						"BenchmarkStringProcessor": false,
						"BenchmarkNumberCruncher":  false,
						"BenchmarkSubProcessor":    false,
					}

					for _, name := range benchmarks {
						if _, exists := expectedNames[name]; exists {
							expectedNames[name] = true
						}
					}

					for name, found := range expectedNames {
						if !found {
							t.Errorf("DiscoverBenchmarks() did not find expected benchmark: %s", name)
						}
					}
				} else {
					// When not expecting benchmarks, verify we got an empty list
					if len(benchmarks) != tt.expectedCount {
						t.Errorf("DiscoverBenchmarks() returned %d benchmarks, want %d", len(benchmarks), tt.expectedCount)
					}
				}
			}
		})
	}
}

func TestDiscoverBenchmarksWithInvalidScope(t *testing.T) {
	// Test with a non-existent directory
	nonExistentDir := "/path/that/does/not/exist"

	benchmarks, err := benchmark.DiscoverBenchmarks(nonExistentDir)

	if err == nil {
		t.Errorf("DiscoverBenchmarks() expected error for non-existent directory but got none")
	}

	if len(benchmarks) > 0 {
		t.Errorf("DiscoverBenchmarks() returned benchmarks for non-existent directory: %v", benchmarks)
	}
}

func TestDiscoverBenchmarksWithNoGoFiles(t *testing.T) {
	// Create a temporary directory with no Go files
	tempDir, err := os.MkdirTemp("", "benchmark_test_no_go")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a regular file (not a Go file)
	regularFile := filepath.Join(tempDir, "regular.txt")
	if err := os.WriteFile(regularFile, []byte("This is not a Go file"), 0644); err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}

	benchmarks, err := benchmark.DiscoverBenchmarks(tempDir)

	if err != nil {
		t.Errorf("DiscoverBenchmarks() unexpected error: %v", err)
	}

	if len(benchmarks) != 0 {
		t.Errorf("DiscoverBenchmarks() returned benchmarks when none should exist: %v", benchmarks)
	}
}

func TestDiscoverBenchmarksWithNoBenchmarks(t *testing.T) {
	// Create a temporary directory with a Go test file but no benchmarks
	tempDir, err := os.MkdirTemp("", "benchmark_test_no_benchmarks")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with no benchmark functions
	testFile := filepath.Join(tempDir, "no_benchmarks_test.go")
	testContent := `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Log("This is a regular test, not a benchmark")
}

func HelperFunction() {
	// This is not a benchmark
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	benchmarks, err := benchmark.DiscoverBenchmarks(tempDir)

	if err != nil {
		t.Errorf("DiscoverBenchmarks() unexpected error: %v", err)
	}

	if len(benchmarks) != 0 {
		t.Errorf("DiscoverBenchmarks() returned benchmarks when none should exist: %v", benchmarks)
	}
}

func TestDiscoverBenchmarksWithMalformedFunctions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "benchmark_test_malformed")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file with malformed benchmark functions
	testFile := filepath.Join(tempDir, "malformed_test.go")
	testContent := `package test

import "testing"

// This is not a valid benchmark function (wrong parameter type)
func BenchmarkWrongParam(t *testing.T) {
	for i := 0; i < 100; i++ {
		_ = i
	}

// This is not a benchmark function (wrong name)
func NotABenchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = i
	}

// This is not a benchmark function (missing parameter)
func BenchmarkMissingParam() {
	for i := 0; i < 100; i++ {
		_ = i
	}
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	benchmarks, err := benchmark.DiscoverBenchmarks(tempDir)

	if err != nil {
		t.Errorf("DiscoverBenchmarks() unexpected error: %v", err)
	}

	// Should not find any valid benchmarks due to malformed syntax
	if len(benchmarks) != 0 {
		t.Errorf("DiscoverBenchmarks() returned benchmarks for malformed functions: %v", benchmarks)
	}
}

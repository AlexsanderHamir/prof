package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"

	"github.com/AlexsanderHamir/prof/parser"
)

func fixtureV2Path(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join("testdata", "testFilesV2", name)
	if _, err := os.Stat(p); err != nil {
		t.Skip("fixture not present:", p)
	}
	return p
}

func TestGetAllFunctionNamesV2(t *testing.T) { //nolint:gocognit,funlen // table-driven subtests
	// Test with CPU profile
	profilePath := fixtureV2Path(t, "BenchmarkGenPool_cpu.out")

	t.Run("no filters", func(t *testing.T) {
		filter := config.FunctionFilter{}
		names, err := parser.GetAllFunctionNamesV2(profilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed: %v", err)
		}

		// Should return some function names
		if len(names) == 0 {
			t.Error("Expected non-empty function names, got empty slice")
		}

		// Check that all names are non-empty
		for i, name := range names {
			if name == "" {
				t.Errorf("Function name at index %d is empty", i)
			}
		}
	})

	t.Run("with include prefixes", func(t *testing.T) {
		filter := config.FunctionFilter{
			IncludePrefixes: []string{"github.com/AlexsanderHamir/GenPool"},
		}
		names, err := parser.GetAllFunctionNamesV2(profilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed: %v", err)
		}

		// Should return some function names that match the prefix
		if len(names) == 0 {
			t.Error("Expected non-empty function names with prefix filter, got empty slice")
		}

		// All returned names should be from functions that match the prefix
		for i, name := range names {
			if name == "" {
				t.Errorf("Function name at index %d is empty", i)
			}
		}
	})

	t.Run("with ignore functions", func(t *testing.T) {
		filter := config.FunctionFilter{
			IgnoreFunctions: []string{"func1", "func2"},
		}
		names, err := parser.GetAllFunctionNamesV2(profilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed: %v", err)
		}

		// Should return some function names
		if len(names) == 0 {
			t.Error("Expected non-empty function names with ignore filter, got empty slice")
		}

		// Check that ignored functions are not in the result
		ignoredSet := make(map[string]struct{})
		for _, ignored := range filter.IgnoreFunctions {
			ignoredSet[ignored] = struct{}{}
		}

		for _, name := range names {
			if _, ignored := ignoredSet[name]; ignored {
				t.Errorf("Ignored function '%s' should not be in results", name)
			}
		}
	})

	t.Run("with both filters", func(t *testing.T) {
		filter := config.FunctionFilter{
			IncludePrefixes: []string{"github.com/AlexsanderHamir/GenPool"},
			IgnoreFunctions: []string{"func1"},
		}
		names, err := parser.GetAllFunctionNamesV2(profilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed: %v", err)
		}

		// Should return some function names
		if len(names) == 0 {
			t.Error("Expected non-empty function names with both filters, got empty slice")
		}

		// Check that ignored functions are not in the result
		ignoredSet := make(map[string]struct{})
		for _, ignored := range filter.IgnoreFunctions {
			ignoredSet[ignored] = struct{}{}
		}

		for _, name := range names {
			if _, ignored := ignoredSet[name]; ignored {
				t.Errorf("Ignored function '%s' should not be in results", name)
			}
		}
	})

	t.Run("with non-existent prefix", func(t *testing.T) {
		filter := config.FunctionFilter{
			IncludePrefixes: []string{"non/existent/prefix"},
		}
		names, err := parser.GetAllFunctionNamesV2(profilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed: %v", err)
		}

		// Should return empty slice when no functions match the prefix
		if len(names) != 0 {
			t.Errorf("Expected empty slice for non-existent prefix, got %d names", len(names))
		}
	})

	t.Run("with memory profile", func(t *testing.T) {
		memoryProfilePath := fixtureV2Path(t, "BenchmarkGenPool_memory.out")
		filter := config.FunctionFilter{}
		names, err := parser.GetAllFunctionNamesV2(memoryProfilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed with memory profile: %v", err)
		}

		// Should return some function names
		if len(names) == 0 {
			t.Error("Expected non-empty function names from memory profile, got empty slice")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		filter := config.FunctionFilter{}
		_, err := parser.GetAllFunctionNamesV2("non_existent_file.out", filter)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("with block profile", func(t *testing.T) {
		blockProfilePath := fixtureV2Path(t, "BenchmarkGenPool_block.out")
		filter := config.FunctionFilter{}
		names, err := parser.GetAllFunctionNamesV2(blockProfilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed with block profile: %v", err)
		}

		// Should return some function names
		if len(names) == 0 {
			t.Error("Expected non-empty function names from block profile, got empty slice")
		}
	})

	t.Run("with mutex profile", func(t *testing.T) {
		mutexProfilePath := fixtureV2Path(t, "BenchmarkGenPool_mutex.out")
		filter := config.FunctionFilter{}
		names, err := parser.GetAllFunctionNamesV2(mutexProfilePath, filter)
		if err != nil {
			t.Fatalf("GetAllFunctionNamesV2() failed with mutex profile: %v", err)
		}

		// Should return some function names
		if len(names) == 0 {
			t.Error("Expected non-empty function names from mutex profile, got empty slice")
		}
	})
}

func TestOrganizeProfileByPackageV2(t *testing.T) {
	// Use existing test profile file
	profilePath := fixtureV2Path(t, "BenchmarkGenPool_cpu.out")

	// Test with empty filter
	filter := config.FunctionFilter{}
	result, err := parser.OrganizeProfileByPackageV2(profilePath, filter)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the result contains expected package names
	if !strings.Contains(result, "github.com/AlexsanderHamir/GenPool") {
		t.Error("Expected result to contain 'github.com/AlexsanderHamir/GenPool' package")
	}

	// Verify the result contains function names
	if !strings.Contains(result, "func1") {
		t.Error("Expected result to contain 'func1' function")
	}

	// Verify subtotals are present
	if !strings.Contains(result, "Subtotal") {
		t.Error("Expected result to contain subtotals")
	}

	// Verify that percentages are displayed
	if !strings.Contains(result, "%") {
		t.Error("Expected result to contain percentage values")
	}

	t.Logf("Generated report:\n%s", result)
}

func TestOrganizeProfileByPackageV2WithFilter(t *testing.T) {
	// Use existing test profile file
	profilePath := fixtureV2Path(t, "BenchmarkGenPool_cpu.out")

	// Test with include prefix filter
	filter := config.FunctionFilter{
		IncludePrefixes: []string{"github.com/AlexsanderHamir/GenPool"},
	}

	result, err := parser.OrganizeProfileByPackageV2(profilePath, filter)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only contain the specified package
	if !strings.Contains(result, "github.com/AlexsanderHamir/GenPool") {
		t.Error("Expected result to contain 'github.com/AlexsanderHamir/GenPool' package when filtered")
	}
}

func TestOrganizeProfileByPackageV2WithIgnoreFunctions(t *testing.T) {
	// Use existing test profile file
	profilePath := fixtureV2Path(t, "BenchmarkGenPool_cpu.out")

	// Test with ignore functions filter
	filter := config.FunctionFilter{
		IgnoreFunctions: []string{"BenchmarkGenPool"},
	}

	result, err := parser.OrganizeProfileByPackageV2(profilePath, filter)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not contain ignored functions
	if strings.Contains(result, "BenchmarkGenPool") {
		t.Error("Expected result to NOT contain 'BenchmarkGenPool' function when ignored")
	}

	// Should still contain other functions
	if !strings.Contains(result, "github.com/AlexsanderHamir/GenPool") {
		t.Error("Expected result to contain 'github.com/AlexsanderHamir/GenPool' package")
	}
}

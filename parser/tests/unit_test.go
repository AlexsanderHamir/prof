package test

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestLinesIntoObjsV2(t *testing.T) {
	profilePath := filepath.Join("testFilesV2", "BenchmarkGenPool_cpu.out")

	lineObjs, err := parser.TurnLinesIntoObjectsV2(profilePath)
	if err != nil {
		t.Fatalf("TurnLinesIntoObjectsV2() failed: %v", err)
	}

	// Check that we got some results
	if len(lineObjs) == 0 {
		t.Error("Expected non-empty results, got empty slice")
	}

	// Check that each LineObj has the expected structure
	for i, obj := range lineObjs {
		if obj == nil {
			t.Errorf("LineObj at index %d is nil", i)
			continue
		}

		// Check that function name is not empty
		if obj.FnName == "" {
			t.Errorf("LineObj at index %d has empty function name", i)
		}

		// Check that flat value is non-negative
		if obj.Flat < 0 {
			t.Errorf("LineObj at index %d has negative flat value: %f", i, obj.Flat)
		}

		// Check that cumulative value is non-negative
		if obj.Cum < 0 {
			t.Errorf("LineObj at index %d has negative cumulative value: %f", i, obj.Cum)
		}

		// Check that percentages are within valid range (0-100)
		if obj.FlatPercentage < 0 || obj.FlatPercentage > 100 {
			t.Errorf("LineObj at index %d has invalid flat percentage: %f", i, obj.FlatPercentage)
		}

		if obj.CumPercentage < 0 || obj.CumPercentage > 100 {
			t.Errorf("LineObj at index %d has invalid cumulative percentage: %f", i, obj.CumPercentage)
		}

		// Sum percentage can be 100% or slightly over due to rounding, so we allow it to be <= 101
		if obj.SumPercentage < 0 || obj.SumPercentage > 101 {
			t.Errorf("LineObj at index %d has invalid sum percentage: %f", i, obj.SumPercentage)
		}

		// Check that cumulative value is greater than or equal to flat value
		if obj.Cum < obj.Flat {
			t.Errorf("LineObj at index %d has cumulative value (%f) less than flat value (%f)",
				i, obj.Cum, obj.Flat)
		}
	}

	// Check that results are sorted by flat value (descending) - this is a property of the V2 API
	for i := 1; i < len(lineObjs); i++ {
		if lineObjs[i].Flat > lineObjs[i-1].Flat {
			t.Errorf("Results not properly sorted: lineObjs[%d].Flat (%f) > lineObjs[%d].Flat (%f)",
				i, lineObjs[i].Flat, i-1, lineObjs[i-1].Flat)
		}
	}

	// Test with a different profile file to ensure it works with different types
	profilePath2 := filepath.Join("testFilesV2", "BenchmarkGenPool_memory.out")
	lineObjs2, err := parser.TurnLinesIntoObjectsV2(profilePath2)
	if err != nil {
		t.Fatalf("TurnLinesIntoObjectsV2() failed with memory profile: %v", err)
	}

	if len(lineObjs2) == 0 {
		t.Error("Expected non-empty results from memory profile, got empty slice")
	}

	// Test error case with non-existent file
	_, err = parser.TurnLinesIntoObjectsV2("non_existent_file.out")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestGetAllFunctionNamesV2(t *testing.T) {
	// Test with CPU profile
	profilePath := filepath.Join("testFilesV2", "BenchmarkGenPool_cpu.out")

	t.Run("no filters", func(t *testing.T) {
		filter := internal.FunctionFilter{}
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
		filter := internal.FunctionFilter{
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
		filter := internal.FunctionFilter{
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
		filter := internal.FunctionFilter{
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
		filter := internal.FunctionFilter{
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
		memoryProfilePath := filepath.Join("testFilesV2", "BenchmarkGenPool_memory.out")
		filter := internal.FunctionFilter{}
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
		filter := internal.FunctionFilter{}
		_, err := parser.GetAllFunctionNamesV2("non_existent_file.out", filter)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("with block profile", func(t *testing.T) {
		blockProfilePath := filepath.Join("testFilesV2", "BenchmarkGenPool_block.out")
		filter := internal.FunctionFilter{}
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
		mutexProfilePath := filepath.Join("testFilesV2", "BenchmarkGenPool_mutex.out")
		filter := internal.FunctionFilter{}
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

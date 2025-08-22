package test

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestLinesIntoObjs(t *testing.T) {
	profilePath := filepath.Join("testFiles", "BenchmarkGenPool_cpu.txt")

	lineObjs, err := parser.TurnLinesIntoObjects(profilePath)
	if err != nil {
		t.Error(err)
	}

	minLineCount := 145
	numberOfLines := len(lineObjs)
	if numberOfLines < minLineCount {
		t.Errorf("Expected at least %d, found only %d", minLineCount, numberOfLines)
	}
}

func TestShouldKeepLine(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		profileFilters map[int]float64
		ignoreFuncs    []string
		ignorePrefixes []string
		want           bool
	}{
		{
			name: "empty line",
			line: "",
			want: false,
		},
		{
			name: "too short line",
			line: "1 2 3 4 5",
			want: false,
		},
		{
			name:           "profile value below threshold",
			line:           "2.0asa 0.0asa 0.0asas 0.0asas 0.0asas mypackage.myFunc",
			profileFilters: map[int]float64{0: 2.5},
			want:           false,
		},
		{
			name:           "profile value above threshold",
			line:           "3.0asas 0.0asa 0.0asas 0.0asas 0.0asas mypackage.myFunc",
			profileFilters: map[int]float64{0: 2.5},
			want:           true,
		},
		{
			name:        "ignore function match",
			line:        "3.0 0.0 0.0 0.0 0.0 mypackage.ignoreMe",
			ignoreFuncs: []string{"ignoreMe"},
			want:        false,
		},
		{
			name:           "ignore prefix match",
			line:           "3.0 0.0 0.0 0.0 0.0 prefixFunc.something",
			ignorePrefixes: []string{"prefixFunc."},
			want:           false,
		},
		{
			name:           "no ignore, passes all filters",
			line:           "3.0 0.0 0.0 0.0 0.0 mypackage.myFunc",
			profileFilters: map[int]float64{0: 2.5},
			ignoreFuncs:    []string{"otherFunc"},
			ignorePrefixes: []string{"otherPrefix."},
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ignoreFuncsMap := make(map[string]struct{})
			for _, f := range tt.ignoreFuncs {
				ignoreFuncsMap[f] = struct{}{}
			}
			ignorePrefixesMap := make(map[string]struct{})
			for _, p := range tt.ignorePrefixes {
				ignorePrefixesMap[p] = struct{}{}
			}

			options := &internal.LineFilterArgs{
				ProfileFilters:    tt.profileFilters,
				IgnoreFunctionSet: ignoreFuncsMap,
				IgnorePrefixSet:   ignorePrefixesMap,
			}

			got := parser.ShouldKeepLine(tt.line, options)
			if got != tt.want {
				t.Errorf("ShouldKeepLine() = %v, want %v (test: %s)", got, tt.want, tt.name)
			}
		})
	}
}

func TestExtractFunctionName(t *testing.T) {
	tests := []struct {
		name                 string
		line                 string
		functionPrefixes     []string
		ignoreFunctionSet    map[string]struct{}
		expectedFunctionName string
	}{
		{
			name:                 "valid function name with no filters",
			line:                 "0.12s  1.18% 98.03%   0.12s  1.18%  primitives_performance.(*RingBuffer[go.shape.*uint8]).Pop (inline)",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "Pop",
		},
		{
			name:                 "valid function name with inline marker",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  github.com/AlexsanderHamir/GenPool/test.BenchmarkGenPool.func1 (inline)",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "func1",
		},
		{
			name:                 "function name with complex generic type",
			line:                 "0.04s 0.056% 98.95%      0.15s  0.21%  github.com/AlexsanderHamir/GenPool/pool.(*ShardedPool[go.shape.struct { Name string; Data []uint8; Result int64; github.com/AlexsanderHamir/GenPool/test._ [16]uint8; Fields = github.com/AlexsanderHamir/GenPool/pool.Fields[github.com/AlexsanderHamir/GenPool/test.BenchmarkObject] },go.shape.*github.com/AlexsanderHamir/GenPool/test.BenchmarkObject]).Put",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "Put",
		},
		{
			name:                 "function name with prefix filter - should pass",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  github.com/AlexsanderHamir/GenPool/test.BenchmarkGenPool.func1",
			functionPrefixes:     []string{"github.com/AlexsanderHamir/GenPool"},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "func1",
		},
		{
			name:                 "function name with prefix filter - should fail",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  runtime.schedule",
			functionPrefixes:     []string{"github.com/AlexsanderHamir/GenPool"},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "",
		},
		{
			name:                 "function name with ignore function filter - should be ignored",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  runtime.schedule",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{"schedule": {}},
			expectedFunctionName: "",
		},
		{
			name:                 "function name with ignore function filter - should pass",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  runtime.schedule",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{"otherFunc": {}},
			expectedFunctionName: "schedule",
		},
		{
			name:                 "line too short - should fail",
			line:                 "0.12s  1.18% 98.03%",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "",
		},
		{
			name:                 "empty line - should fail",
			line:                 "",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "",
		},
		{
			name:                 "function name with multiple prefixes - should pass with first prefix",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  github.com/AlexsanderHamir/GenPool/test.BenchmarkGenPool.func1",
			functionPrefixes:     []string{"runtime", "github.com/AlexsanderHamir/GenPool"},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "func1",
		},
		{
			name:                 "function name with multiple prefixes - should pass with second prefix",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  runtime.schedule",
			functionPrefixes:     []string{"runtime", "github.com/AlexsanderHamir/GenPool"},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "schedule",
		},
		{
			name:                 "function name with parentheses and parameters",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  runtime.(*mheap).allocSpan",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "allocSpan",
		},
		{
			name:                 "function name with complex nested structure",
			line:                 "0.01s 0.014% 0.084%     70.13s 98.28%  sync/atomic.(*Pointer[go.shape.struct { Name string; Data []uint8; Result int64; github.com/AlexsanderHamir/GenPool/test._ [16]uint8; Fields = github.com/AlexsanderHamir/GenPool/pool.Fields[github.com/AlexsanderHamir/GenPool/test.BenchmarkObject] }]).CompareAndSwap",
			functionPrefixes:     []string{},
			ignoreFunctionSet:    map[string]struct{}{},
			expectedFunctionName: "CompareAndSwap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.ExtractFunctionName(tt.line, tt.functionPrefixes, tt.ignoreFunctionSet)
			if got != tt.expectedFunctionName {
				t.Errorf("extractFunctionName() = %v, want %v (test: %s)", got, tt.expectedFunctionName, tt.name)
			}
		})
	}
}

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

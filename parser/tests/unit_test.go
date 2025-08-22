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

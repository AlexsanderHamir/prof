package test

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestLinesIntoObjs(t *testing.T) {
	profilePath := filepath.Join("testFiles", "BenchmarkGenPool_cpu.txt")
	profileType := "cpu"

	lineObjs, err := parser.TurnLinesIntoObjects(profilePath, profileType)
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

			options := &args.LineFilterArgs{
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

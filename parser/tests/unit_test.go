package test

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/parser"
)

func TestLinesIntoObjs(t *testing.T) {
	profilePath := filepath.Join("testFiles", "BenchmarkGenPool_cpu.txt")
	profileType := "cpu"

	lineObjs, err := parser.TurnLinesIntoObjects(profilePath, profileType)
	if err != nil {
		t.Error(err)
	}

	minLineCount := 50
	numberOfLines := len(lineObjs)
	if numberOfLines < minLineCount {
		t.Errorf("Expected at least %d, found only %d", minLineCount, numberOfLines)
	}
}

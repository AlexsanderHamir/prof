package benchmark

import (
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tooling"
)

func TestRunBenchmarks_nilRunner(t *testing.T) {
	err := RunBenchmarks(nil, []string{"B"}, []string{"cpu"}, "t", 1, false, false, false)
	if err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("got %v", err)
	}
}

func TestRunBenchmarks_emptyBenchmarks(t *testing.T) {
	err := RunBenchmarks(tooling.NewExecRunner(), nil, []string{"cpu"}, "t", 1, false, false, false)
	if err == nil || !strings.Contains(err.Error(), "benchmarks flag is empty") {
		t.Fatalf("got %v", err)
	}
}

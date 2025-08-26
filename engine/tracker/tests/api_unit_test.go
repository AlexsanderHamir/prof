package tests_test

import (
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tracker"
)

func TestTrackAuto(t *testing.T) {
	tagPath1 := "tag1"
	tagPath2 := "tag2"
	benchName := "BenchmarkGenPool"
	profileTypes := []string{"memory", "cpu", "mutex", "block"}

	for _, profileType := range profileTypes {
		t.Run(profileType, func(t *testing.T) {
			selections := tracker.Selections{
				Baseline:      tagPath1,
				Current:       tagPath2,
				BenchmarkName: benchName,
				ProfileType:   profileType,
			}

			profileResult, err := tracker.CheckPerformanceDifferences(&selections)
			if err != nil {
				t.Fatal(err)
			}

			if profileResult == nil {
				t.Fatal("profileResult should not be nil")
			}

			first := profileResult.FunctionChanges[0]
			if first == nil {
				t.Fatal("first report should not be nil")
			}

			report := first.Report()
			if report == "" {
				t.Fatalf("report is missing")
			}

			if profileResult == nil {
				t.Fatal("nil result")
			}
		})
	}
}

func TestTrackManual(t *testing.T) {
	filePath1 := "bench/tag1/bin/BenchmarkGenPool/BenchmarkGenPool_cpu.out"
	filePath2 := "bench/tag2/bin/BenchmarkGenPool/BenchmarkGenPool_cpu.out"

	selections := tracker.Selections{
		Baseline: filePath1,
		Current:  filePath2,
		IsManual: true,
	}

	profileResult, err := tracker.CheckPerformanceDifferences(&selections)
	if err != nil {
		t.Fatal(err)
	}

	if profileResult == nil {
		t.Fatal("profileResult should not be nil")
	}

	first := profileResult.FunctionChanges[0]
	if first == nil {
		t.Fatal("first report should not be nil")
	}

	report := first.Report()
	if report == "" {
		t.Fatalf("report is missing")
	}

	if profileResult == nil {
		t.Fatal("nil result")
	}
}

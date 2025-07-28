package tests_test

import (
	"testing"

	"github.com/AlexsanderHamir/prof/tracker"
)

func TestCoreBlock(t *testing.T) {
	tagPath1 := "tag1"
	tagPath2 := "tag2"
	benchName := "BenchmarkGenPool"
	profileTypes := []string{"memory", "cpu", "mutex", "block"}

	for _, profileType := range profileTypes {
		t.Run(profileType, func(t *testing.T) {
			profileResult, err := tracker.CheckPerformanceDifferences(tagPath1, tagPath2, benchName, profileType)
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

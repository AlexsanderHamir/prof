package tests

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/tracker"
)

func TestCoreBlock(t *testing.T) {
	tagPath1 := filepath.Join("bench", "tag1")
	tagPath2 := filepath.Join("bench", "tag2")
	benchName := "BenchmarkGenPool"
	profileTypes := []string{"memory", "cpu", "mutex", "block"}

	for _, profileType := range profileTypes {
		t.Run(profileType, func(t *testing.T) {
			profileResult, err := tracker.CheckPerformanceDifferences(tagPath1, tagPath2, benchName, profileType)
			if err != nil {
				t.Fatal(err)
			}

			if profileResult == nil {
				t.Fatal("nil result")
			}
		})
	}
}

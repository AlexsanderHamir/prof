package tests

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/regressor"
)

func TestCoreBlock(t *testing.T) {
	tagPath1 := filepath.Join("bench", "tag1")
	tagPath2 := filepath.Join("bench", "tag2")
	benchName := "BenchmarkGenPool"
	profileType := "cpu"

	err := regressor.CheckPerformanceDifferences(tagPath1, tagPath2, benchName, profileType)
	if err != nil {
		t.Fatal(err)
	}
}

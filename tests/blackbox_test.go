package tests

import (
	"testing"
)

func TestConfig(t *testing.T) {
	originalValue := envDirName
	withCleanUp := true

	label := "WithFunctionFilter"
	t.Run(label, func(t *testing.T) {
		withconfig := true
		expectedFiles := map[string]bool{
			"BenchmarkStringProcessor.txt": false,
			"ProcessStrings.txt":           false,
			"GenerateStrings.txt":          false,
			"AddString.txt":                false,
		}

		testConfigScenario(t, withconfig, withCleanUp, label, originalValue, expectedFiles)
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		withConfig := false
		var expectedProfiles map[string]bool // nil

		testConfigScenario(t, withConfig, withCleanUp, label, originalValue, expectedProfiles)
	})
}

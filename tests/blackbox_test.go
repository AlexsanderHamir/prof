package tests

import (
	"testing"

	"github.com/AlexsanderHamir/prof/config"
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

		cfg := &config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IncludePrefixes: []string{"test-environment"},
				},
			},
		}

		testConfigScenario(t, cfg, withconfig, withCleanUp, label, originalValue, expectedFiles)
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		withConfig := false

		var expectedProfiles map[string]bool // empty
		var cfg config.Config                // empty

		testConfigScenario(t, &cfg, withConfig, withCleanUp, label, originalValue, expectedProfiles)
	})
}

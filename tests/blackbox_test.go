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

		expectedFiles := map[fileFullName]IsFileExpected{
			"BenchmarkStringProcessor.txt": IsFileExpected(true),
			"ProcessStrings.txt":           IsFileExpected(true),
			"GenerateStrings.txt":          IsFileExpected(true),
			"AddString.txt":                IsFileExpected(true),
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

		var expectedProfiles map[fileFullName]IsFileExpected // empty
		var cfg config.Config                                // empty

		testConfigScenario(t, &cfg, withConfig, withCleanUp, label, originalValue, expectedProfiles)
	})
}

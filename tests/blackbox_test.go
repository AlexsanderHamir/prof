package tests

import (
	"testing"

	"github.com/AlexsanderHamir/prof/config"
)

func TestConfig(t *testing.T) {
	withCleanUp := true

	label := "WithFunctionFilter"
	t.Run(label, func(t *testing.T) {
		isExpected := IsFileExpected(true)
		specifiedFiles := map[fileFullName]IsFileExpected{
			"BenchmarkStringProcessor.txt":        isExpected,
			"ProcessStrings.txt":                  isExpected,
			"GenerateStrings.txt":                 isExpected,
			"AddString.txt":                       isExpected,
			"BenchmarkStringProcessor_cpu.png":    isExpected,
			"BenchmarkStringProcessor_memory.png": isExpected,
		}

		cfg := &config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IncludePrefixes: []string{"test-environment"},
				},
			},
		}

		withconfig := true
		expectNonSpecifiedFiles := false
		testConfigScenario(t, cfg, expectNonSpecifiedFiles, withconfig, withCleanUp, label, specifiedFiles)
		if !withCleanUp {
			envDirName = envDirNameStatic
		}
	})

	label = "WithFunctionIgnore"
	t.Run(label, func(t *testing.T) {
		isNotExpected := IsFileExpected(false)
		speficiedFiles := map[fileFullName]IsFileExpected{
			"GenerateStrings.txt":          IsFileExpected(true),
			"BenchmarkStringProcessor.txt": isNotExpected,
			"ProcessStrings.txt":           isNotExpected,
			"AddString.txt":                isNotExpected,
		}

		cfg := &config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IgnoreFunctions: []string{"BenchmarkStringProcessor", "ProcessStrings", "AddString"},
				},
			},
		}

		withconfig := true
		expectNonSpecifiedFiles := true
		testConfigScenario(t, cfg, expectNonSpecifiedFiles, withconfig, withCleanUp, label, speficiedFiles)
		if !withCleanUp {
			envDirName = envDirNameStatic
		}
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		var specifiedFiles map[fileFullName]IsFileExpected // empty
		var cfg config.Config                              // empty

		withConfig := false
		expectNonSpecifiedFiles := true
		testConfigScenario(t, &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, label, specifiedFiles)
		if !withCleanUp {
			envDirName = envDirNameStatic
		}
	})
}

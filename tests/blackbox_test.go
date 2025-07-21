package tests

import (
	"testing"

	"github.com/AlexsanderHamir/prof/config"
)

func TestConfig(t *testing.T) {
	withCleanUp := false

	label := "WithFunctionFilter"
	t.Run(label, func(t *testing.T) {
		specifiedFiles := map[fileFullName]*FieldsCheck{
			"BenchmarkStringProcessor.txt":        newDefaultFieldsCheckExpected(),
			"ProcessStrings.txt":                  newDefaultFieldsCheckExpected(),
			"GenerateStrings.txt":                 newDefaultFieldsCheckExpected(),
			"AddString.txt":                       newDefaultFieldsCheckExpected(),
			"BenchmarkStringProcessor_cpu.png":    newDefaultFieldsCheckExpected(),
			"BenchmarkStringProcessor_memory.png": newDefaultFieldsCheckExpected(),
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
	})

	label = "WithFunctionIgnore"
	t.Run(label, func(t *testing.T) {

		speficiedFiles := map[fileFullName]*FieldsCheck{
			"GenerateStrings.txt":          newDefaultFieldsCheckExpected(),
			"BenchmarkStringProcessor.txt": newDefaultFieldsCheckNotExpected(),
			"ProcessStrings.txt":           newDefaultFieldsCheckNotExpected(),
			"AddString.txt":                newDefaultFieldsCheckNotExpected(),
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
	})

	label = "WithFunctionFilterPlusIgnore"
	t.Run(label, func(t *testing.T) {

		specifiedFiles := map[fileFullName]*FieldsCheck{
			"BenchmarkStringProcessor_cpu.png":    newDefaultFieldsCheckExpected(),
			"BenchmarkStringProcessor_memory.png": newDefaultFieldsCheckExpected(),
			"GenerateStrings.txt":                 newDefaultFieldsCheckExpected(),
			"BenchmarkStringProcessor.txt":        newDefaultFieldsCheckNotExpected(),
			"ProcessStrings.txt":                  newDefaultFieldsCheckNotExpected(),
			"AddString.txt":                       newDefaultFieldsCheckNotExpected(),
		}

		cfg := &config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IncludePrefixes: []string{"test-environment"},
					IgnoreFunctions: []string{"BenchmarkStringProcessor", "ProcessStrings", "AddString"},
				},
			},
		}

		withconfig := true
		expectNonSpecifiedFiles := false
		testConfigScenario(t, cfg, expectNonSpecifiedFiles, withconfig, withCleanUp, label, specifiedFiles)
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		var specifiedFiles map[fileFullName]*FieldsCheck // empty
		var cfg config.Config                            // empty

		withConfig := false
		expectNonSpecifiedFiles := true
		testConfigScenario(t, &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, label, specifiedFiles)
	})
}

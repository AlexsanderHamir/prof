package tests

import (
	"fmt"
	"testing"

	"github.com/AlexsanderHamir/prof/config"
)

func TestConfig(t *testing.T) {
	withCleanUp := true

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

		cfg := config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IncludePrefixes: []string{"test-environment"},
				},
			},
		}

		withConfig := true
		expectNonSpecifiedFiles := false
		noConfigFile := false
		cmd := defaultCmd()
		testConfigScenario(t, "", &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, noConfigFile, label, specifiedFiles, cmd)
	})

	label = "WithFunctionIgnore"
	t.Run(label, func(t *testing.T) {

		specifiedFiles := map[fileFullName]*FieldsCheck{
			"GenerateStrings.txt":          newDefaultFieldsCheckExpected(),
			"BenchmarkStringProcessor.txt": newDefaultFieldsCheckNotExpected(),
			"ProcessStrings.txt":           newDefaultFieldsCheckNotExpected(),
			"AddString.txt":                newDefaultFieldsCheckNotExpected(),
		}

		cfg := config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IgnoreFunctions: []string{"BenchmarkStringProcessor", "ProcessStrings", "AddString"},
				},
			},
		}

		withConfig := true
		expectNonSpecifiedFiles := true
		noConfigFile := false
		cmd := defaultCmd()
		testConfigScenario(t, "", &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, noConfigFile, label, specifiedFiles, cmd)
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

		cfg := config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IncludePrefixes: []string{"test-environment"},
					IgnoreFunctions: []string{"BenchmarkStringProcessor", "ProcessStrings", "AddString"},
				},
			},
		}

		withConfig := true
		expectNonSpecifiedFiles := false

		noConfigFile := false
		cmd := defaultCmd()
		testConfigScenario(t, "", &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, noConfigFile, label, specifiedFiles, cmd)
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		var specifiedFiles map[fileFullName]*FieldsCheck // empty
		var cfg config.Config                            // empty

		withConfig := false
		expectNonSpecifiedFiles := true
		noConfigFile := false
		cmd := defaultCmd()
		testConfigScenario(t, "", &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, noConfigFile, label, specifiedFiles, cmd)
	})

	label = "WithoutConfigFile"
	t.Run(label, func(t *testing.T) {
		var specifiedFiles map[fileFullName]*FieldsCheck // empty
		var cfg config.Config                            // empty

		withConfig := false
		expectNonSpecifiedFiles := true
		noConfigFile := true
		cmd := defaultCmd()
		testConfigScenario(t, "", &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, noConfigFile, label, specifiedFiles, cmd)
	})
}

func TestProfileValidation(t *testing.T) {
	withCleanUp := true

	label := "RandomProfileName"
	t.Run(label, func(t *testing.T) {
		var specifiedFiles map[fileFullName]*FieldsCheck // empty
		var cfg config.Config                            // empty

		withConfig := false
		expectNonSpecifiedFiles := true
		noConfigFile := true
		cmd := []string{
			"--benchmarks", fmt.Sprintf("[%s]", benchName),
			"--profiles", fmt.Sprintf("[%s,%s,%s]", cpuProfile, memProfile, "fakeProfileName"),
			"--count", count,
			"--tag", tag,
		}
		expectedErrorMessage := "failed to run BenchmarkStringProcessor: profile fakeProfileName is not supported"
		testConfigScenario(t, expectedErrorMessage, &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, noConfigFile, label, specifiedFiles, cmd)
	})

	label = "NonCollectedProfile"
	t.Run(label, func(t *testing.T) {
		var specifiedFiles map[fileFullName]*FieldsCheck // empty
		var cfg config.Config                            // empty

		withConfig := false
		expectNonSpecifiedFiles := true
		noConfigFile := true
		cmd := []string{
			"--benchmarks", fmt.Sprintf("[%s]", benchName),
			"--profiles", fmt.Sprintf("[%s,%s,%s]", cpuProfile, memProfile, goroutineProfile),
			"--count", count,
			"--tag", tag,
		}
		expectedErrorMessage := fmt.Sprintf("failed to run %s: flag provided but not defined: -%s", benchName, goroutineProfile)
		testConfigScenario(t, expectedErrorMessage, &cfg, expectNonSpecifiedFiles, withConfig, withCleanUp, noConfigFile, label, specifiedFiles, cmd)
	})
}

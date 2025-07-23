package tests

import (
	"fmt"
	"testing"

	"github.com/AlexsanderHamir/prof/config"
)

func TestConfig(t *testing.T) {
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

		testArgs := &TestArgs{
			specifiedFiles:          specifiedFiles,
			cfg:                     cfg,
			withConfig:              true,
			expectNonSpecifiedFiles: false,
			noConfigFile:            false,
			cmd:                     defaultCmd(),
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile},
		}

		testConfigScenario(t, testArgs)
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

		testArgs := &TestArgs{
			specifiedFiles:          specifiedFiles,
			cfg:                     cfg,
			withConfig:              true,
			expectNonSpecifiedFiles: true,
			noConfigFile:            false,
			cmd:                     defaultCmd(),
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile},
		}

		testConfigScenario(t, testArgs)
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

		testArgs := &TestArgs{
			specifiedFiles:          specifiedFiles,
			cfg:                     cfg,
			withConfig:              true,
			expectNonSpecifiedFiles: false,
			noConfigFile:            false,
			cmd:                     defaultCmd(),
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile},
		}

		testConfigScenario(t, testArgs)
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            false,
			cmd:                     defaultCmd(),
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile},
		}

		testConfigScenario(t, testArgs)
	})

	label = "WithoutConfigFile"
	t.Run(label, func(t *testing.T) {
		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     defaultCmd(),
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile},
		}

		testConfigScenario(t, testArgs)
	})
}

func TestProfileValidation(t *testing.T) {
	label := "RandomProfileName"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			"--benchmarks", fmt.Sprintf("[%s]", benchName),
			"--profiles", fmt.Sprintf("[%s,%s,%s]", cpuProfile, memProfile, "fakeProfileName"),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "failed to parse benchmark config: received unvalid profile",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})

	label = "NonCollectedProfile"
	t.Run(label, func(t *testing.T) {
		nonCollectedProfile := "goroutine"
		cmd := []string{
			"--benchmarks", fmt.Sprintf("[%s]", benchName),
			"--profiles", fmt.Sprintf("[%s]", nonCollectedProfile),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "failed to parse benchmark config: received unvalid profile: goroutine",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})

	label = "CollectedProfile"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			"--benchmarks", fmt.Sprintf("[%s]", benchName),
			"--profiles", fmt.Sprintf("[%s,%s,%s]", cpuProfile, memProfile, blockProfile),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   4, // cpu, mem, goroutine/trace, block
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile, blockProfile},
		}

		testConfigScenario(t, testArgs)
	})
}

func TestCommandValidation(t *testing.T) {
	label := "EmptyBenchmarkSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			"--benchmarks", "[]",
			"--profiles", fmt.Sprintf("[%s,%s]", cpuProfile, memProfile),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "failed to parse benchmark config: benchmarks argument cannot be an empty list",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})

	label = "NoBracketBenchmarkSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			"--benchmarks", benchName,
			"--profiles", fmt.Sprintf("[%s,%s]", cpuProfile, memProfile),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "failed to parse benchmark config: benchmarks argument must be wrapped in brackets",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})

	label = "NoBracketProfileSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			"--benchmarks", fmt.Sprintf("[%s]", benchName),
			"--profiles", cpuProfile, memProfile,
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    `Error: unknown command "memory" for "prof"`,
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})

	label = "EmptyBracketProfileSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			"--benchmarks", fmt.Sprintf("[%s]", benchName),
			"--profiles", "[]",
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     config.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "failed to parse benchmark config: profiles argument cannot be an empty list",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})
}

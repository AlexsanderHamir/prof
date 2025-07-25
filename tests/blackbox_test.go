package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/shared"
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
			cmd:                     defaultRunCmd(),
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
			cmd:                     defaultRunCmd(),
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
			cmd:                     defaultRunCmd(),
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
			cmd:                     defaultRunCmd(),
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
			cmd:                     defaultRunCmd(),
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
		profileName := "fakeProfileName"
		cmd := []string{
			"run",
			"--benchmarks", benchName,
			"--profiles", fmt.Sprintf("%s,%s,%s", cpuProfile, memProfile, profileName),
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
			expectedErrorMessage:    fmt.Sprintf("failed to run %s: profile %s is not supported", benchName, profileName),
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})

	label = "NonCollectedProfile"
	t.Run(label, func(t *testing.T) {
		profileName := "goroutine"
		cmd := []string{
			"run",
			"--benchmarks", benchName,
			"--profiles", profileName,
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
			expectedErrorMessage:    fmt.Sprintf("failed to run %s: profile %s is not supported", benchName, profileName),
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
			"run",
			"--benchmarks", benchName,
			"--profiles", fmt.Sprintf("%s,%s,%s", cpuProfile, memProfile, blockProfile),
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
			expectedNumberOfFiles:   4, // cpu, mem, goroutine, block
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
			"run",
			"--benchmarks", "",
			"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
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
			expectedErrorMessage:    "benchmarks flag is empty",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})

	label = "EmptyProfileSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			"run",
			"--benchmarks", benchName,
			"--profiles", "",
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
			expectedErrorMessage:    "profiles flag is empty",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
		}

		testConfigScenario(t, testArgs)
	})
}

func TestManualCommand(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}
	binaryPath := path.Join(root, testDirName, "prof")

	buildProf(t, binaryPath, root)

	t.Cleanup(func() {
		benchPath := path.Join(root, testDirName, shared.MainDirOutput)
		if err = os.RemoveAll(benchPath); err != nil {
			t.Logf("Failed to clean up bench: %v", err)
		}

		if err = os.Remove("prof"); err != nil {
			t.Logf("failed to clean prof binary: %s", err)
		}
	})

	label := "BasicRun"
	t.Run(label, func(t *testing.T) {
		args := []string{
			"manual",
			"--tag", tag,
			"assets/cpu.out",
			"assets/memory.out",
			"assets/block.out",
			"assets/mutex.out",
		}

		cmd := exec.Command("./prof", args...)
		cmd.Dir = path.Join(root, testDirName)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			t.Error(err)
		}

		if stdout.Len() > 0 {
			fmt.Println(stdout.String())
		}
	})
}

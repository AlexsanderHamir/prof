package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
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

		cfg := internal.Config{
			FunctionFilter: map[string]internal.FunctionFilter{
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
			checkSuccessMessage:     true,
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

		cfg := internal.Config{
			FunctionFilter: map[string]internal.FunctionFilter{
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
			checkSuccessMessage:     true,
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

		cfg := internal.Config{
			FunctionFilter: map[string]internal.FunctionFilter{
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
			checkSuccessMessage:     true,
		}

		testConfigScenario(t, testArgs)
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     internal.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            false,
			cmd:                     defaultRunCmd(),
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile},
			checkSuccessMessage:     true,
		}

		testConfigScenario(t, testArgs)
	})

	label = "WithoutConfigFile"
	t.Run(label, func(t *testing.T) {
		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     internal.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     defaultRunCmd(),
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile},
			checkSuccessMessage:     true,
		}

		testConfigScenario(t, testArgs)
	})
}

func TestProfileValidation(t *testing.T) {
	label := "RandomProfileName"
	t.Run(label, func(t *testing.T) {
		profileName := "fakeProfileName"
		cmd := []string{
			internal.AUTOCMD,
			"--benchmarks", benchName,
			"--profiles", fmt.Sprintf("%s,%s,%s", cpuProfile, memProfile, profileName),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     internal.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    fmt.Sprintf("failed to run %s: profile %s is not supported", benchName, profileName),
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
			checkSuccessMessage:     true,
		}

		testConfigScenario(t, testArgs)
	})

	label = "NonCollectedProfile"
	t.Run(label, func(t *testing.T) {
		profileName := "goroutine"
		cmd := []string{
			internal.AUTOCMD,
			"--benchmarks", benchName,
			"--profiles", profileName,
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     internal.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    fmt.Sprintf("failed to run %s: profile %s is not supported", benchName, profileName),
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
			checkSuccessMessage:     true,
		}

		testConfigScenario(t, testArgs)
	})

	label = "CollectedProfile"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			internal.AUTOCMD,
			"--benchmarks", benchName,
			"--profiles", fmt.Sprintf("%s,%s,%s", cpuProfile, memProfile, blockProfile),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     internal.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "",
			label:                   label,
			expectedNumberOfFiles:   4, // cpu, mem, goroutine, block
			withCleanUp:             true,
			expectedProfiles:        []string{cpuProfile, memProfile, blockProfile},
			checkSuccessMessage:     true,
		}

		testConfigScenario(t, testArgs)
	})
}

func TestCommandValidation(t *testing.T) {
	label := "EmptyBenchmarkSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			internal.AUTOCMD,
			"--benchmarks", "",
			"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     internal.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "benchmarks flag is empty",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
			checkSuccessMessage:     true,
		}

		testConfigScenario(t, testArgs)
	})

	label = "EmptyProfileSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			internal.AUTOCMD,
			"--benchmarks", benchName,
			"--profiles", "",
			"--count", count,
			"--tag", tag,
		}

		testArgs := &TestArgs{
			specifiedFiles:          nil,
			cfg:                     internal.Config{},
			withConfig:              false,
			expectNonSpecifiedFiles: true,
			noConfigFile:            true,
			cmd:                     cmd,
			expectedErrorMessage:    "profiles flag is empty",
			label:                   label,
			expectedNumberOfFiles:   3,
			withCleanUp:             true,
			expectedProfiles:        nil,
			checkSuccessMessage:     true,
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
		benchPath := path.Join(root, testDirName, internal.MainDirOutput)
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
			internal.MANUALCMD,
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

// TestTrackerBasicRun is not inspecting the results, but just ensuring that no errors occur when the command is run.
func TestTrackerBasicRun(t *testing.T) {
	// 1. Set up
	label := "testing"
	runs := "5"
	tagName := "tag1"
	blockOutputCheck := true
	isEnvironmentSet := false
	checkSuccessMessage := false
	createBenchForTracker(t, label, runs, tagName, blockOutputCheck, isEnvironmentSet)
	envDirName = "Enviroment"

	runs = "10"
	tagName = "tag2"
	isEnvironmentSet = true
	createBenchForTracker(t, label, runs, tagName, blockOutputCheck, isEnvironmentSet)

	root, err := getProjectRoot()
	if err != nil {
		t.Error(err)
	}

	envFullPath := path.Join(root, testDirName, "Enviroment testing")
	t.Cleanup(func() {
		if err = os.RemoveAll(envFullPath); err != nil {
			t.Logf("Failed to clean up bench: %v", err)
		}
	})

	// 2. Test Tracker
	label = "Auto"
	t.Run(label, func(t *testing.T) {
		args := []string{
			"track",
			internal.AUTOCMD,
			"--base", "tag1",
			"--current", "tag2",
			"--bench-name", benchName,
			"--profile-type", "cpu",
			"--output-format", "summary",
		}

		shouldContinue := runProf(t, envFullPath, args, "", checkSuccessMessage)
		if !shouldContinue {
			t.Error("runProf failed")
		}
	})

	label = "Manual"
	baseTag := "bench/tag1/text/BenchmarkStringProcessor/BenchmarkStringProcessor_cpu.txt"
	Current := "bench/tag2/text/BenchmarkStringProcessor/BenchmarkStringProcessor_cpu.txt"
	outputFormat := "summary"
	t.Run(label, func(t *testing.T) {
		args := []string{
			"track",
			internal.MANUALCMD,
			"--base", baseTag,
			"--current", Current,
			"--output-format", outputFormat,
		}

		shouldContinue := runProf(t, envFullPath, args, "", checkSuccessMessage)
		if !shouldContinue {
			t.Error("runProf failed")
		}
	})
}

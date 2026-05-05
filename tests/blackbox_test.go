package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

func skipSlowIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("slow integration scenario; omit -short to run full integration suite")
	}
}

func TestConfig(t *testing.T) { //nolint:funlen // scenario matrix
	skipSlowIntegration(t)
	label := "WithFunctionFilter"
	t.Run(label, func(t *testing.T) {
		specifiedFiles := map[fileFullName]*FieldsCheck{
			"BenchmarkStringProcessor.txt": newDefaultFieldsCheckExpected(),
			"ProcessStrings.txt":           newDefaultFieldsCheckExpected(),
			"GenerateStrings.txt":          newDefaultFieldsCheckExpected(),
			"AddString.txt":                newDefaultFieldsCheckExpected(),
		}

		cfg := internal.Config{
			FunctionFilter: map[string]internal.FunctionFilter{
				benchName: {
					// Prefer symbols under the synthetic module when present; also match package-qualified names when paths are trimmed (Linux CI).
					IncludePrefixes: []string{"test-environment", "utils."},
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
			"GenerateStrings.txt":          newDefaultFieldsCheckExpected(),
			"BenchmarkStringProcessor.txt": newDefaultFieldsCheckNotExpected(),
			"ProcessStrings.txt":           newDefaultFieldsCheckNotExpected(),
			"AddString.txt":                newDefaultFieldsCheckNotExpected(),
		}

		cfg := internal.Config{
			FunctionFilter: map[string]internal.FunctionFilter{
				benchName: {
					IncludePrefixes: []string{"test-environment", "utils."},
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
	skipSlowIntegration(t)
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
		cmd = append(cmd, autoBenchSkipPNGArgs()...)

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
		cmd = append(cmd, autoBenchSkipPNGArgs()...)

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
		cmd = append(cmd, autoBenchSkipPNGArgs()...)

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
	skipSlowIntegration(t)
	label := "EmptyBenchmarkSlice"
	t.Run(label, func(t *testing.T) {
		cmd := []string{
			internal.AUTOCMD,
			"--benchmarks", "",
			"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
			"--count", count,
			"--tag", tag,
		}
		cmd = append(cmd, autoBenchSkipPNGArgs()...)

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
		cmd = append(cmd, autoBenchSkipPNGArgs()...)

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
	skipSlowIntegration(t)
	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}
	binaryPath := filepath.Join(root, testDirName, profBinaryName())

	buildProf(t, binaryPath, root)

	t.Cleanup(func() {
		benchPath := filepath.Join(root, testDirName, internal.MainDirOutput)
		if err = os.RemoveAll(benchPath); err != nil {
			t.Logf("Failed to clean up bench: %v", err)
		}

		if err = os.Remove(filepath.Join(root, testDirName, profBinaryName())); err != nil {
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

		cmd := exec.Command(binaryPath, args...)
		cmd.Dir = filepath.Join(root, testDirName)

		var stderr bytes.Buffer
		cmd.Stdout = io.Discard
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			t.Error(err)
		}

		// Manual collect logs per-function progress (collector), not InfoCollectionSuccess (auto benchmark pipeline only).
		if !strings.Contains(stderr.String(), "Collected function") {
			t.Fatalf("expected stderr to contain collector progress; stderr=%q", stderr.String())
		}

		benchRoot := filepath.Join(root, testDirName, internal.MainDirOutput, tag)
		if fi, statErr := os.Stat(benchRoot); statErr != nil || !fi.IsDir() {
			t.Fatalf("expected bench output directory %s: %v", benchRoot, statErr)
		}
	})
}

func TestTrackerBasicRun(t *testing.T) {
	skipSlowIntegration(t)
	// 1. Set up
	label := "testing"
	runs := "1"
	tagName := "tag1"
	blockOutputCheck := true
	isEnvironmentSet := false
	checkSuccessMessage := false
	createBenchForTracker(t, label, runs, tagName, blockOutputCheck, isEnvironmentSet)

	runs = "1"
	tagName = "tag2"
	isEnvironmentSet = true
	createBenchForTracker(t, label, runs, tagName, blockOutputCheck, isEnvironmentSet)

	root, err := getProjectRoot()
	if err != nil {
		t.Error(err)
	}

	envFullPath := filepath.Join(root, testDirName, integrationEnvDir(label))
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

		stdout, stderr, ok := runProfCaptured(t, envFullPath, args, "", checkSuccessMessage)
		if !ok {
			t.Fatal("runProf failed")
		}
		combined := stdout + stderr
		if !strings.Contains(stdout, "Performance Tracking Summary") &&
			!strings.Contains(stdout, "Total Functions Analyzed") &&
			!strings.Contains(combined, "No function changes detected") {
			t.Fatalf("unexpected track output; stdout=%q stderr=%q", stdout, stderr)
		}
	})

	label = "Manual"
	baseTag := "bench/tag1/bin/BenchmarkStringProcessor/BenchmarkStringProcessor_cpu.out"
	currentProfile := "bench/tag2/bin/BenchmarkStringProcessor/BenchmarkStringProcessor_cpu.out"
	outputFormat := "summary"
	t.Run(label, func(t *testing.T) {
		args := []string{
			"track",
			internal.MANUALCMD,
			"--base", baseTag,
			"--current", currentProfile,
			"--output-format", outputFormat,
		}

		stdout, stderr, ok := runProfCaptured(t, envFullPath, args, "", checkSuccessMessage)
		if !ok {
			t.Fatal("runProf failed")
		}
		combined := stdout + stderr
		if !strings.Contains(stdout, "Performance Tracking Summary") &&
			!strings.Contains(stdout, "Total Functions Analyzed") &&
			!strings.Contains(combined, "No function changes detected") {
			t.Fatalf("unexpected track output; stdout=%q stderr=%q", stdout, stderr)
		}
	})
}

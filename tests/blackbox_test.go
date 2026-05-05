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

// TestAutoEndToEnd validates the full `prof auto` pipeline once: build the
// prof binary, run `go test -bench`, collect profiles, write the bench/<tag>/
// layout. Runs at smokeCount because we're asserting wiring + layout, not
// CPU sampling stability — filter behavior is covered by TestFunctionFilter
// against deterministic committed fixtures.
func TestAutoEndToEnd(t *testing.T) {
	skipSlowIntegration(t)

	label := "Smoke"
	testArgs := &TestArgs{
		cfg:                     internal.Config{},
		withConfig:              false,
		expectNonSpecifiedFiles: true,
		noConfigFile:            true,
		cmd:                     runCmdWithCount(smokeCount),
		label:                   label,
		expectedNumberOfFiles:   3,
		withCleanUp:             true,
		expectedProfiles:        []string{cpuProfile, memProfile},
		checkSuccessMessage:     true,
	}
	testConfigScenario(t, testArgs)
}

// TestFunctionFilter exercises every FunctionFilter combination the previous
// TestConfig matrix covered, but in-process against committed pprof fixtures.
// No subprocess, no `go test -bench`, no per-scenario synthetic Go module —
// each subtest is a pure call into parser.GetAllFunctionNamesV2 +
// collector.GetFunctionsOutput followed by directory assertions.
func TestFunctionFilter(t *testing.T) {
	cases := []struct {
		name          string
		cfg           internal.Config
		expected      map[fileFullName]*FieldsCheck
		expectNonSpec bool
	}{
		{"WithFunctionFilter", configWithFilter(), expectAllFunctionFiles(), false},
		{"WithFunctionIgnore", configWithIgnore(), expectOnlyGenerate(), true},
		{"WithFunctionFilterPlusIgnore", configWithFilterAndIgnore(), expectOnlyGenerate(), false},
		{"WithoutAnyConfig", internal.Config{}, nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			perProfileNames := runFilterInProcess(t, tc.cfg)
			for _, names := range perProfileNames {
				checkFilteredNamesAgainstSpec(t, names, tc.expected, tc.expectNonSpec)
			}
		})
	}
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

	label = "Auto"
	t.Run(label, func(t *testing.T) {
		args := []string{
			"track",
			internal.AUTOCMD,
			"--base", "tag1",
			"--current", "tag2",
			"--bench-name", benchName,
			"--profile-type", cpuProfile,
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
	baseTag := "bench/tag1/bin/" + benchName + "/" + benchName + "_" + cpuProfile + ".out"
	currentProfile := "bench/tag2/bin/" + benchName + "/" + benchName + "_" + cpuProfile + ".out"
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

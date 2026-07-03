package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/cli"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func skipSlowIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("slow integration scenario; omit -short to run full integration suite (see TESTING.md)")
	}
}

// TestAutoEndToEnd validates the full `prof auto` pipeline once: build the
// prof binary, run `go test -bench`, collect profiles, write the bench/<tag>/
// layout. Runs at smokeCount because we're asserting wiring + layout, not
// CPU sampling stability — filter behavior is covered by TestFunctionFilter
// against deterministic committed fixtures.
func TestAutoEndToEnd(t *testing.T) {
	skipSlowIntegration(t)

	testArgs := &TestArgs{
		cfg:                     config.Config{},
		withConfig:              false,
		expectNonSpecifiedFiles: true,
		cmd:                     runCmdWithCount(smokeCount),
		expectedNumberOfFiles:   3,
		expectedProfiles:        []string{cpuProfile, memProfile},
		checkSuccessMessage:     true,
		useSharedEnv:            true,
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
		cfg           config.Config
		expected      map[fileFullName]*FieldsCheck
		expectNonSpec bool
	}{
		{"WithFunctionFilter", configWithFilter(), expectAllFunctionFiles(), false},
		{"WithFunctionIgnore", configWithIgnore(), expectOnlyGenerate(), true},
		{"WithFunctionFilterPlusIgnore", configWithFilterAndIgnore(), expectOnlyGenerate(), false},
		{"WithoutAnyConfig", config.Config{}, nil, true},
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
	t.Run("RandomProfileName", func(t *testing.T) {
		profileName := "fakeProfileName"
		cmd := []string{
			cli.CmdAuto,
			"--benchmarks", benchName,
			"--profiles", fmt.Sprintf("%s,%s,%s", cpuProfile, memProfile, profileName),
			"--count", validationCount,
			"--tag", tag,
		}

		testConfigScenario(t, &TestArgs{
			cfg:                  config.Config{},
			cmd:                  cmd,
			expectedErrorMessage: fmt.Sprintf("failed to run %s: profile %s is not supported", benchName, profileName),
			checkSuccessMessage:  true,
			useSharedEnv:         true,
		})
	})

	t.Run("NonCollectedProfile", func(t *testing.T) {
		profileName := "goroutine"
		cmd := []string{
			cli.CmdAuto,
			"--benchmarks", benchName,
			"--profiles", profileName,
			"--count", validationCount,
			"--tag", tag,
		}

		testConfigScenario(t, &TestArgs{
			cfg:                  config.Config{},
			cmd:                  cmd,
			expectedErrorMessage: fmt.Sprintf("failed to run %s: profile %s is not supported", benchName, profileName),
			checkSuccessMessage:  true,
			useSharedEnv:         true,
		})
	})

	t.Run("CollectedProfile", func(t *testing.T) {
		cmd := []string{
			cli.CmdAuto,
			"--benchmarks", benchName,
			"--profiles", fmt.Sprintf("%s,%s,%s", cpuProfile, memProfile, blockProfile),
			"--count", validationCount,
			"--tag", tag,
		}

		testConfigScenario(t, &TestArgs{
			cfg:                     config.Config{},
			expectNonSpecifiedFiles: true,
			cmd:                     cmd,
			expectedNumberOfFiles:   4, // cpu, mem, goroutine, block
			expectedProfiles:        []string{cpuProfile, memProfile, blockProfile},
			checkSuccessMessage:     true,
			useSharedEnv:            true,
		})
	})
}

func TestCommandValidation(t *testing.T) {
	skipSlowIntegration(t)
	t.Run("EmptyBenchmarkSlice", func(t *testing.T) {
		cmd := []string{
			cli.CmdAuto,
			"--benchmarks", "",
			"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
			"--count", validationCount,
			"--tag", tag,
		}

		testConfigScenario(t, &TestArgs{
			cfg:                  config.Config{},
			cmd:                  cmd,
			expectedErrorMessage: "benchmarks flag is empty",
			checkSuccessMessage:  true,
			useSharedEnv:         true,
		})
	})

	t.Run("EmptyProfileSlice", func(t *testing.T) {
		cmd := []string{
			cli.CmdAuto,
			"--benchmarks", benchName,
			"--profiles", "",
			"--count", validationCount,
			"--tag", tag,
		}

		testConfigScenario(t, &TestArgs{
			cfg:                  config.Config{},
			cmd:                  cmd,
			expectedErrorMessage: "profiles flag is empty",
			checkSuccessMessage:  true,
			useSharedEnv:         true,
		})
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
		benchPath := filepath.Join(root, workspace.MainDirOutput, tag)
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
			cli.CmdManual,
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

		benchRoot := filepath.Join(root, workspace.MainDirOutput, tag)
		if fi, statErr := os.Stat(benchRoot); statErr != nil || !fi.IsDir() {
			t.Fatalf("expected bench output directory %s: %v", benchRoot, statErr)
		}
	})
}

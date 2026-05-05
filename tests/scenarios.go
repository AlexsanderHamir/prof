package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

// TestArgs holds inputs and expectations for a single integration test scenario.
type TestArgs struct {
	specifiedFiles          map[fileFullName]*FieldsCheck
	cfg                     internal.Config
	expectedNumberOfFiles   int
	expectedErrorMessage    string
	label                   string
	cmd                     []string
	expectedProfiles        []string
	withConfig              bool
	expectNonSpecifiedFiles bool
	noConfigFile            bool
	withCleanUp             bool
	blockOutputCheck        bool
	isEnvironmentSet        bool
	checkSuccessMessage     bool
}

func createConfigFile(t *testing.T, envDir string, cfgTemplate *internal.Config) {
	t.Helper()

	configPath := filepath.Join(envDir, templateFile)

	data, err := json.MarshalIndent(cfgTemplate, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal config template: %v", err)
	}

	if err = os.WriteFile(configPath, data, internal.PermFile); err != nil {
		t.Fatalf("failed to write config template file: %v", err)
	}
}

func testConfigScenario(t *testing.T, testArgs *TestArgs) {
	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}

	envDir := integrationEnvDir(testArgs.label)
	envFullPath := filepath.Join(root, testDirName, envDir)

	if testArgs.withCleanUp {
		t.Cleanup(func() {
			if err = os.RemoveAll(envFullPath); err != nil {
				t.Logf("Failed to clean up environment: %v", err)
			}
		})
	}

	if !testArgs.isEnvironmentSet {
		setupEnviroment(t, envDir)

		if !testArgs.noConfigFile {
			createConfigFile(t, envDir, &testArgs.cfg)
		}

		setUpProf(t, root, envDir)
	}

	shouldContinue := runProf(t, envFullPath, testArgs.cmd, testArgs.expectedErrorMessage, testArgs.checkSuccessMessage)
	if !shouldContinue {
		return
	}

	if !testArgs.blockOutputCheck {
		checkOutput(t, envFullPath, testArgs)
	}
}

func defaultRunCmd() []string {
	cmd := []string{
		internal.AUTOCMD,
		"--benchmarks", benchName,
		"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
		"--count", count,
		"--tag", tag,
	}
	return append(cmd, autoBenchSkipPNGArgs()...)
}

// autoBenchSkipPNGArgs avoids requiring Graphviz during integration tests while `prof auto`
// defaults to strict PNG generation.
func autoBenchSkipPNGArgs() []string {
	return []string{"--skip-png"}
}

func createBenchForTracker(t *testing.T, label, iterations, tagName string, blockOutputCheck, isEnvironmentSet bool) {
	cmd := []string{
		internal.AUTOCMD,
		"--benchmarks", benchName,
		"--profiles", cpuProfile,
		"--count", iterations,
		"--tag", tagName,
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
		expectedNumberOfFiles:   3,
		withCleanUp:             false,
		expectedProfiles:        nil,
		blockOutputCheck:        blockOutputCheck,
		isEnvironmentSet:        isEnvironmentSet,
	}

	testConfigScenario(t, testArgs)
}

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

// runCmdWithCount returns the canonical `prof auto` argv for the synthetic
// benchmark, parameterized by --count so callers can pick smokeCount,
// validationCount, or count without rebuilding the rest of the command.
func runCmdWithCount(countVal string) []string {
	cmd := []string{
		internal.AUTOCMD,
		"--benchmarks", benchName,
		"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
		"--count", countVal,
		"--tag", tag,
	}
	return append(cmd, autoBenchSkipPNGArgs()...)
}

func defaultRunCmd() []string {
	return runCmdWithCount(count)
}

// configWithFilter builds the FunctionFilter scenario that whitelists symbols
// under the synthetic module + utils package prefixes.
func configWithFilter() internal.Config {
	return internal.Config{
		FunctionFilter: map[string]internal.FunctionFilter{
			benchName: {IncludePrefixes: filterIncludePrefixes},
		},
	}
}

// configWithIgnore builds the FunctionFilter scenario that drops the
// canonical ignore set (Benchmark + ProcessStrings + AddString).
func configWithIgnore() internal.Config {
	return internal.Config{
		FunctionFilter: map[string]internal.FunctionFilter{
			benchName: {IgnoreFunctions: filterIgnoreFunctions},
		},
	}
}

// configWithFilterAndIgnore combines the two filter axes; only GenerateStrings
// survives the include + ignore intersection.
func configWithFilterAndIgnore() internal.Config {
	return internal.Config{
		FunctionFilter: map[string]internal.FunctionFilter{
			benchName: {
				IncludePrefixes: filterIncludePrefixes,
				IgnoreFunctions: filterIgnoreFunctions,
			},
		},
	}
}

// expectAllFunctionFiles marks every entry in expectedFunctionFiles as
// expected — used when the filter whitelists the whole synthetic module.
func expectAllFunctionFiles() map[fileFullName]*FieldsCheck {
	m := make(map[fileFullName]*FieldsCheck, len(expectedFunctionFiles))
	for _, name := range expectedFunctionFiles {
		m[fileFullName(name)] = newDefaultFieldsCheckExpected()
	}
	return m
}

// expectOnlyGenerate marks GenerateStrings as expected and the other three
// canonical functions as filtered-out — used by both ignore-only and
// filter+ignore scenarios.
func expectOnlyGenerate() map[fileFullName]*FieldsCheck {
	return map[fileFullName]*FieldsCheck{
		functionFile(funcGenerate):  newDefaultFieldsCheckExpected(),
		functionFile(benchName):     newDefaultFieldsCheckNotExpected(),
		functionFile(funcProcess):   newDefaultFieldsCheckNotExpected(),
		functionFile(funcAddString): newDefaultFieldsCheckNotExpected(),
	}
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

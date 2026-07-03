package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/cli"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// TestArgs holds inputs and expectations for a single integration test scenario.
type TestArgs struct {
	specifiedFiles          map[fileFullName]*FieldsCheck
	cfg                     config.Config
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
	// useSharedEnv routes the scenario to ensureSharedEnv() instead of
	// building a per-label env under tests/Enviroment <label>/. The shared
	// env is created on demand and cleaned up by TestMain. Setup-related
	// flags (label, isEnvironmentSet, withCleanUp, noConfigFile) are
	// ignored when useSharedEnv is true.
	useSharedEnv bool
}

func createConfigFile(t *testing.T, envDir string, cfgTemplate *config.Config) {
	t.Helper()

	configPath := filepath.Join(envDir, templateFile)

	data, err := json.MarshalIndent(cfgTemplate, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal config template: %v", err)
	}

	if err = os.WriteFile(configPath, data, workspace.PermFile); err != nil {
		t.Fatalf("failed to write config template file: %v", err)
	}
}

func testConfigScenario(t *testing.T, testArgs *TestArgs) {
	envFullPath := resolveScenarioEnv(t, testArgs)

	shouldContinue := runProf(t, envFullPath, testArgs.cmd, testArgs.expectedErrorMessage, testArgs.checkSuccessMessage)
	if !shouldContinue {
		return
	}

	if !testArgs.blockOutputCheck {
		checkOutput(t, envFullPath, testArgs)
	}
}

// resolveScenarioEnv returns the env path the scenario should run in, either
// from the shared env or a per-label one. Per-label envs are built on first
// use and (when withCleanUp is set) torn down by t.Cleanup.
func resolveScenarioEnv(t *testing.T, testArgs *TestArgs) string {
	t.Helper()

	if testArgs.useSharedEnv {
		return ensureSharedEnv(t)
	}

	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}

	envDir := integrationEnvDir(testArgs.label)
	envFullPath := filepath.Join(root, testDirName, envDir)

	if testArgs.withCleanUp {
		t.Cleanup(func() {
			if cleanupErr := os.RemoveAll(envFullPath); cleanupErr != nil {
				t.Logf("Failed to clean up environment: %v", cleanupErr)
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

	return envFullPath
}

// runCmdWithCount returns the canonical `prof auto` argv for the synthetic
// benchmark, parameterized by --count so callers can pick smokeCount or
// validationCount without rebuilding the rest of the command.
func runCmdWithCount(countVal string) []string {
	cmd := []string{
		cli.CmdAuto,
		"--benchmarks", benchName,
		"--profiles", fmt.Sprintf("%s,%s", cpuProfile, memProfile),
		"--count", countVal,
		"--tag", tag,
	}
	return append(cmd, autoBenchSkipPNGArgs()...)
}

// configWithFilter builds the collection filter scenario that whitelists symbols
// under the synthetic module + utils package prefixes.
func configWithFilter() config.Config {
	return config.Config{
		Version: config.CurrentVersion,
		Collection: config.Collection{
			Benchmarks: map[string]config.FunctionFilter{
				benchName: {IncludePrefixes: filterIncludePrefixes},
			},
		},
	}
}

// configWithIgnore builds the collection filter scenario that drops the
// canonical ignore set (Benchmark + ProcessStrings + AddString).
func configWithIgnore() config.Config {
	return config.Config{
		Version: config.CurrentVersion,
		Collection: config.Collection{
			Benchmarks: map[string]config.FunctionFilter{
				benchName: {IgnoreFunctions: filterIgnoreFunctions},
			},
		},
	}
}

// configWithFilterAndIgnore combines the two filter axes; only GenerateStrings
// survives the include + ignore intersection.
func configWithFilterAndIgnore() config.Config {
	return config.Config{
		Version: config.CurrentVersion,
		Collection: config.Collection{
			Benchmarks: map[string]config.FunctionFilter{
				benchName: {
					IncludePrefixes: filterIncludePrefixes,
					IgnoreFunctions: filterIgnoreFunctions,
				},
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

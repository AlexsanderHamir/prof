package tests

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/AlexsanderHamir/prof/config"
)

func TestConfig(t *testing.T) {
	originalValue := envDirName
	withCleanUp := true

	label := "WithConfigFunctionFilter"
	t.Run(label, func(t *testing.T) {
		withconfig := true
		testConfigScenario(t, withconfig, withCleanUp, label, originalValue)
	})

	label = "WithoutAnyConfig"
	t.Run(label, func(t *testing.T) {
		withConfig := false
		testConfigScenario(t, withConfig, withCleanUp, label, originalValue)
	})
}

func testConfigScenario(t *testing.T, withConfig, withCleanUp bool, label, originalValue string) {
	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}

	envDirName = envDirName + " " + label
	envPath := path.Join(root, testDirName, envDirName)

	if withCleanUp {
		t.Cleanup(func() {
			if err := os.RemoveAll(envPath); err != nil {
				t.Logf("Failed to clean up environment: %v", err)
			}
		})
	}

	// 1. Set up Environment
	setupEnviroment(t)

	// 2. Create config conditionally
	var cfg config.Config // empty config
	if withConfig {
		cfg = config.Config{
			FunctionFilter: map[string]config.FunctionFilter{
				benchName: {
					IncludePrefixes: []string{"test-environment"},
				},
			},
		}
	}

	createConfigFile(t, &cfg)

	// 3. Build prof and move to Environment
	setUpProf(t, root)

	// 4. Run ./prof inside the Environment
	args := []string{
		"--benchmarks", fmt.Sprintf("[%s]", benchName),
		"--profiles", fmt.Sprintf("[%s,%s]", cpuProfile, memProfile),
		"--count", count,
		"--tag", tag,
	}
	runProf(t, root, args)

	// 5. Check bench output
	expectedProfiles := []string{cpuProfile, memProfile}
	checkOutput(t, envPath, expectedProfiles, withConfig)

	envDirName = originalValue
}

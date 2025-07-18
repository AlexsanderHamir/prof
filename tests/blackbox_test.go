package tests

import (
	"os"
	"path"
	"testing"

	"github.com/AlexsanderHamir/prof/config"
)

// TestProfNoConfig contains a config file, but an empty one.
func TestProfNoConfig(t *testing.T) {
	root, err := getProjectRoot()
	if err != nil {
		t.Log(err)
	}

	// 1. Set up Enviroment
	setupEnviroment(t)

	// 2. Create specific config inside the Enviroment directory
	emptyConfigFile := &config.Config{}
	createConfigFile(t, emptyConfigFile)

	// 3. Build prof and move to Enviroment
	setUpProf(t, root)

	// 4. Run ./prof inside the Enviroment
	runProf(t, root, []string{
		"--benchmarks", "[BenchmarkStringProcessor]",
		"--profiles", "[cpu,memory]",
		"--count", "3",
		"--tag", "test",
	})

	fullPath := path.Join(root, testDirName, envDirName)
	defer func() {
		if err := os.RemoveAll(fullPath); err != nil {
			t.Logf("Failed to clean up environment: %v", err)
		}
	}()
}

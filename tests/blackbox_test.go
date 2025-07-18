package tests

import (
	"fmt"
	"path"
	"testing"

	"github.com/AlexsanderHamir/prof/config"
)

// TestNoConfig contains a config file, but an empty one.
func TestNoConfig(t *testing.T) {
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
		"--benchmarks", fmt.Sprintf("[%s]", benchName),
		"--profiles", fmt.Sprintf("[%s,%s]", cpuProfile, memProfile),
		"--count", count,
		"--tag", tag,
	})

	envPath := path.Join(root, testDirName, envDirName)

	// 5. Check bench output
	checkOutput(t, envPath, []string{cpuProfile, memProfile})

	// 6. Clean up
	// defer func() {
	// 	if err := os.RemoveAll(envPath); err != nil {
	// 		t.Logf("Failed to clean up environment: %v", err)
	// 	}
	// }()
}

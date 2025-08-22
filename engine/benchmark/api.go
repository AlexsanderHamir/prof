package benchmark

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/prof/internal"
)

func RunBenchmarks(benchmarks, profiles []string, tag string, count int) error {
	if len(benchmarks) == 0 {
		return errors.New("benchmarks flag is empty")
	}

	if len(profiles) == 0 {
		return errors.New("profiles flag is empty")
	}

	cfg, err := internal.LoadFromFile(internal.ConfigFilename)
	if err != nil {
		slog.Info("No config file found at repository root; proceeding without function filters.", "expected", internal.ConfigFilename)
		slog.Info("You can generate one with 'prof setup'. It will be placed at the root next to go.mod.")
		cfg = &internal.Config{}
	}

	if err = setupDirectories(tag, benchmarks, profiles); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	benchArgs := &internal.BenchArgs{
		Benchmarks: benchmarks,
		Profiles:   profiles,
		Count:      count,
		Tag:        tag,
	}

	internal.PrintConfiguration(benchArgs, cfg.FunctionFilter)

	if err = runBenchAndGetProfiles(benchArgs, cfg.FunctionFilter); err != nil {
		return err
	}

	return nil
}

// DiscoverBenchmarks scans the Go module for benchmark functions and returns their names.
// A benchmark is identified by functions matching:
//
//	func BenchmarkXxx(b *testing.B) { ... }
//
// If scope is provided, only searches within that directory and its subdirectories.
// If scope is empty, searches the entire module from the root.
func DiscoverBenchmarks(scope string) ([]string, error) {
	var searchRoot string
	var err error
	
	if scope != "" {
		// Use the provided scope directory
		searchRoot = scope
	} else {
		// Fall back to searching from module root
		searchRoot, err = internal.FindGoModuleRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to locate module root: %w", err)
		}
	}

	return scanForBenchmarks(searchRoot)
}

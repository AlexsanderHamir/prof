package benchmark

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/prof/internal"
)

// RunBenchmarks validates flags, loads optional repo config, prepares bench layout, then runs the full pipeline.
func RunBenchmarks(benchmarks, profiles []string, tag string, count int, groupByPackage bool, lenientProfiles bool, skipPNG bool) error {
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

	if err = runBenchAndGetProfiles(benchArgs, cfg.FunctionFilter, groupByPackage, lenientProfiles, skipPNG); err != nil {
		return err
	}
	return nil
}

// DiscoverBenchmarks scans for functions matching:
//
//	func BenchmarkXxx(b *testing.B) { ... }
//
// If scope is non-empty, search starts there; otherwise the module root is used.
func DiscoverBenchmarks(scope string) ([]string, error) {
	var searchRoot string
	var err error
	if scope != "" {
		searchRoot = scope
	} else {
		searchRoot, err = internal.FindGoModuleRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to locate module root: %w", err)
		}
	}
	return scanForBenchmarks(searchRoot)
}

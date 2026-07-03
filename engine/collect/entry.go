package collect

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// RunAuto validates flags, loads optional repo config, prepares bench layout, then runs the full pipeline.
func RunAuto(runner tooling.Runner, opts AutoOptions) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	if len(opts.Benchmarks) == 0 {
		return errors.New("benchmarks flag is empty")
	}
	if len(opts.Profiles) == 0 {
		return errors.New("profiles flag is empty")
	}
	if opts.Count < 1 {
		return errors.New("count must be at least 1")
	}

	session := termui.NewSession(os.Stderr, int(os.Stderr.Fd()))
	autoSkippedPNG := applyAutoSkipPNG(&opts)

	cfg, err := config.Load()
	if err != nil {
		if session.Interactive() {
			session.Warn("No prof.json found; proceeding without function filters (run prof config init to add one).")
		} else {
			slog.Info("No config file found at repository root; proceeding without function filters.", "expected", config.Filename)
			slog.Info("You can generate one with 'prof config init' or Create prof.json in prof ui.")
		}
		cfg = &config.Config{}
	}

	if err = setupDirectories(opts.Tag, opts.Benchmarks, opts.Profiles, session.Interactive()); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	autoArgs := &config.AutoArgs{
		Benchmarks: opts.Benchmarks,
		Profiles:   opts.Profiles,
		Count:      opts.Count,
		Tag:        opts.Tag,
	}

	if !session.Interactive() {
		config.PrintAutoConfiguration(autoArgs, cfg)
	}

	if autoSkippedPNG {
		if session.Interactive() {
			session.Warn(tooling.SkipPNGNotice)
		} else {
			fmt.Fprintln(os.Stdout, tooling.SkipPNGNotice)
			slog.Info(tooling.SkipPNGNotice)
		}
	}

	return runBenchAndGetProfiles(runner, autoArgs, cfg, opts.LenientProfiles, opts.SkipPNG, session)
}

// applyAutoSkipPNG enables SkipPNG when Graphviz is unavailable. Returns true if it changed opts.
func applyAutoSkipPNG(opts *AutoOptions) bool {
	if opts.SkipPNG || tooling.GraphvizAvailable() {
		return false
	}
	opts.SkipPNG = true
	return true
}

// DiscoverBenchmarks scans for BenchmarkXxx functions under scope or module root.
func DiscoverBenchmarks(scope string) ([]string, error) {
	var searchRoot string
	var err error
	if scope != "" {
		searchRoot = scope
	} else {
		searchRoot, err = workspace.FindModuleRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to locate module root: %w", err)
		}
	}
	return scanForBenchmarks(searchRoot)
}

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/cli"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/tracker"
	"github.com/AlexsanderHamir/prof/version"
)

const configFilePath = "config_template.json"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n🚨🚨🚨 FATAL ERROR 🚨🚨🚨\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintf(os.Stderr, "🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨\n\n")
		os.Exit(1)
	}
}

// run parses arguments and dispatches to the appropriate handler.
func run() error {
	cliArgs, err := cli.ParseArguments()
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	if cliArgs.Version {
		return handleVersion()
	}

	if cliArgs.Command == "setup" {
		return handleSetup(cliArgs)
	}

	if cliArgs.Command == "tracker" {
		return handleTracker(cliArgs)
	}

	return handleBenchmarks(cliArgs)
}

func handleTracker(cliAgrs *cli.Arguments) error {
	base := cliAgrs.BaseLineTagName
	curr := cliAgrs.CurrentTagName
	bench := cliAgrs.BenchmarkName
	profileType := cliAgrs.ProfileType

	res, err := tracker.CheckPerformanceDifferences(base, curr, bench, profileType)
	if err != nil {
		return fmt.Errorf("tracker failed: %w", err)
	}

	fmt.Println(res.FunctionChanges[0].Report())

	return nil
}

// handleVersion prints the current and latest version information.
func handleVersion() error {
	current, latest := version.Check()
	output := version.FormatOutput(current, latest)
	slog.Info(strings.TrimSpace(output))
	return nil
}

// handleSetup processes the setup command and creates a template if requested.
func handleSetup(cliArgs *cli.Arguments) error {
	if cliArgs.CreateTemplate {
		return config.CreateTemplate(cliArgs.OutputPath)
	}
	return errors.New("setup command requires --create-template flag")
}

// handleBenchmarks runs the benchmark pipeline based on parsed arguments.
func handleBenchmarks(cliArgs *cli.Arguments) error {
	if !cli.ValidateRequiredArgs(cliArgs) {
		return errors.New("missing required arguments")
	}

	cfg, err := config.LoadFromFile(configFilePath)
	if err != nil {
		cfg = &config.Config{} // default
	}

	benchmarks, profiles, err := cli.ParseBenchmarkConfig(cliArgs.Benchmarks, cliArgs.Profiles)
	if err != nil {
		return fmt.Errorf("failed to parse benchmark config: %w", err)
	}

	if err = cli.SetupDirectories(cliArgs.Tag, benchmarks, profiles); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	benchArgs := &args.BenchArgs{
		Benchmarks: benchmarks,
		Profiles:   profiles,
		Count:      cliArgs.Count,
		Tag:        cliArgs.Tag,
	}

	cli.PrintConfiguration(benchArgs, cfg.FunctionFilter)

	if err = cli.RunBencAndGetProfiles(benchArgs, cfg.FunctionFilter); err != nil {
		return err
	}

	if cliArgs.GeneralAnalyze {
		if err = cli.AnalyzeProfiles(cliArgs.Tag, profiles, cfg, cliArgs.FlagProfiles); err != nil {
			return fmt.Errorf("failed to analyze profiles: %w", err)
		}
	}

	if cliArgs.FlagProfiles {
		if err = cli.AnalyzeProfiles(cliArgs.Tag, profiles, cfg, cliArgs.FlagProfiles); err != nil {
			return fmt.Errorf("failed to flag profiles: %w", err)
		}
	}

	return nil
}

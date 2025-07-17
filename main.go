package main

import (
	"fmt"
	"log"

	"github.com/AlexsanderHamir/prof/cli"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/version"
)

const configFilePath = "config_template.json"

// main is the entry point for the prof tool.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run parses arguments and dispatches to the appropriate handler.
func run() error {
	args, err := cli.ParseArguments()
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	if args.Version {
		return handleVersion()
	}

	if args.Command == "setup" {
		return handleSetup(args)
	}

	return handleBenchmarks(args)
}

// handleVersion prints the current and latest version information.
func handleVersion() error {
	current, latest := version.Check()
	fmt.Print(version.FormatOutput(current, latest))
	return nil
}

// handleSetup processes the setup command and creates a template if requested.
func handleSetup(args *cli.Arguments) error {
	if args.CreateTemplate {
		return config.CreateTemplate(args.OutputPath)
	}
	return fmt.Errorf("setup command requires --create-template flag")
}

// handleBenchmarks runs the benchmark pipeline based on parsed arguments.
func handleBenchmarks(args *cli.Arguments) error {
	if !cli.ValidateRequiredArgs(args) {
		return fmt.Errorf("missing required arguments")
	}

	cfg, err := config.LoadFromFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	benchmarks, profiles, err := cli.ParseBenchmarkConfig(args.Benchmarks, args.Profiles)
	if err != nil {
		return fmt.Errorf("failed to parse benchmark config: %w", err)
	}

	if err := cli.SetupDirectories(args.Tag, benchmarks, profiles); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	cli.PrintConfiguration(benchmarks, profiles, args.Tag, args.Count, cfg.BenchmarkConfigs)

	if err := cli.RunBenchmarksAndProcessProfiles(benchmarks, profiles, args.Count, args.Tag, cfg.BenchmarkConfigs); err != nil {
		return fmt.Errorf("failed to run benchmarks: %w", err)
	}

	if args.GeneralAnalyze {
		if err := cli.AnalyzeProfiles(args.Tag, profiles, cfg, args.FlagProfiles); err != nil {
			return fmt.Errorf("failed to analyze profiles: %w", err)
		}
	}

	if args.FlagProfiles {
		if err := cli.AnalyzeProfiles(args.Tag, profiles, cfg, args.FlagProfiles); err != nil {
			return fmt.Errorf("failed to flag profiles: %w", err)
		}
	}

	return nil
}

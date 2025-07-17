package main

import (
	"fmt"
	"log"

	"github.com/AlexsanderHamir/prof/cli"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/version"
)


func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

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

func handleVersion() error {
	current, latest := version.Check()
	fmt.Print(version.FormatOutput(current, latest))
	return nil
}

func handleSetup(args *cli.Arguments) error {
	if args.CreateTemplate {
		return config.CreateTemplate(args.OutputPath)
	}
	return fmt.Errorf("setup command requires --create-template flag")
}

func handleBenchmarks(args *cli.Arguments) error {
	if !cli.ValidateRequiredArgs(args) {
		return fmt.Errorf("missing required arguments")
	}

	cfg, err := config.LoadFromFile("config_template.json")
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
		if err := cli.AnalyzeProfiles(args.Tag, profiles, cfg, false); err != nil {
			return fmt.Errorf("failed to analyze profiles: %w", err)
		}
	}

	if args.FlagProfiles {
		if err := cli.AnalyzeProfiles(args.Tag, profiles, cfg, true); err != nil {
			return fmt.Errorf("failed to flag profiles: %w", err)
		}
	}

	return nil
}

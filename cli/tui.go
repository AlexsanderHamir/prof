package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/engine/benchmark"
	"github.com/AlexsanderHamir/prof/internal/args"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/shared"
	"github.com/spf13/cobra"
)

const (
	tuiPageSize = 20
)

func createTuiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive selection of benchmarks and profiles, then runs prof auto",
		RunE:  runTUI,
	}

	return cmd
}

func runTUI(_ *cobra.Command, _ []string) error {
	benchNames, err := benchmark.DiscoverBenchmarks()
	if err != nil {
		return fmt.Errorf("failed to discover benchmarks: %w", err)
	}
	if len(benchNames) == 0 {
		return errors.New("no benchmarks found in this module (look for func BenchmarkXxx(b *testing.B) in *_test.go)")
	}

	// Select benchmarks
	var selectedBenches []string
	benchPrompt := &survey.MultiSelect{
		Message:  "Select benchmarks to run:",
		Options:  benchNames,
		PageSize: tuiPageSize,
	}
	if err = survey.AskOne(benchPrompt, &selectedBenches, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Select profiles
	profilesOptions := []string{"cpu", "memory", "mutex", "block"}
	var selectedProfiles []string
	profilesPrompt := &survey.MultiSelect{
		Message: "Select profiles:",
		Options: profilesOptions,
		Default: []string{"cpu"},
	}
	if err = survey.AskOne(profilesPrompt, &selectedProfiles, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Count
	var countStr string
	countPrompt := &survey.Input{Message: "Number of runs (count):", Default: "1"}
	if err = survey.AskOne(countPrompt, &countStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	runCount, convErr := strconv.Atoi(countStr)
	if convErr != nil || runCount < 1 {
		return fmt.Errorf("invalid count: %s", countStr)
	}

	// Tag
	var tagStr string
	tagPrompt := &survey.Input{Message: "Tag name (used to group results under bench/<tag>):"}
	if err = survey.AskOne(tagPrompt, &tagStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	cfg, err := config.LoadFromFile(shared.ConfigFilename)
	if err != nil {
		// Not fatal; proceed without function filters
		cfg = &config.Config{}
	}

	if err = benchmark.SetupDirectories(tagStr, selectedBenches, selectedProfiles); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	benchArgs := &args.BenchArgs{
		Benchmarks: selectedBenches,
		Profiles:   selectedProfiles,
		Count:      runCount,
		Tag:        tagStr,
	}

	printConfiguration(benchArgs, cfg.FunctionFilter)
	if err = runBenchAndGetProfiles(benchArgs, cfg.FunctionFilter); err != nil {
		return err
	}

	return nil
}

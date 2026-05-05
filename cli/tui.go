package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

func runTUI(svc *app.Services, _ *cobra.Command, _ []string) error {
	// Get current working directory for scope-aware benchmark discovery
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	benchNames, err := svc.Benchmark.DiscoverBenchmarks(currentDir)
	if err != nil {
		return fmt.Errorf("failed to discover benchmarks: %w", err)
	}

	if len(benchNames) == 0 {
		return errors.New("no benchmarks found in this directory or its subdirectories (look for func BenchmarkXxx(b *testing.B) in *_test.go)")
	}

	var selectedBenches []string
	benchPrompt := &survey.MultiSelect{
		Message:  "Select benchmarks to run:",
		Options:  benchNames,
		PageSize: tuiPageSize,
	}
	if err = survey.AskOne(benchPrompt, &selectedBenches, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	profilesOptions := svc.Benchmark.SupportedProfiles()
	var selectedProfiles []string
	profilesPrompt := &survey.MultiSelect{
		Message: "Select profiles:",
		Options: profilesOptions,
		Default: []string{"cpu"},
	}

	if err = survey.AskOne(profilesPrompt, &selectedProfiles, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	var countStr string
	countPrompt := &survey.Input{Message: "Number of runs (count):", Default: "1"}
	if err = survey.AskOne(countPrompt, &countStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	runCount, convErr := strconv.Atoi(countStr)
	if convErr != nil || runCount < 1 {
		return fmt.Errorf("invalid count: %s", countStr)
	}

	var tagStr string
	tagPrompt := &survey.Input{Message: "Tag name (used to group results under bench/<tag>):"}
	if err = survey.AskOne(tagPrompt, &tagStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	var groupPkg bool
	if err = survey.AskOne(&survey.Confirm{
		Message: "Group profile output by Go package (writes additional *_grouped.txt style reports under the tag)?",
		Default: false,
	}, &groupPkg); err != nil {
		return err
	}

	var lenient bool
	if err = survey.AskOne(&survey.Confirm{
		Message: "Lenient profiles: if a profile binary is missing after the bench run, skip it instead of failing?",
		Default: false,
	}, &lenient); err != nil {
		return err
	}

	var skipPng bool
	if err = survey.AskOne(&survey.Confirm{
		Message: "Skip PNG generation failures (e.g. when Graphviz is not installed)? The run still succeeds if text profiles were produced.",
		Default: false,
	}, &skipPng); err != nil {
		return err
	}

	if err = svc.Benchmark.RunBenchmarks(selectedBenches, selectedProfiles, tagStr, runCount, groupPkg, lenient, skipPng); err != nil {
		return err
	}

	return nil
}

func runTUITrackAuto(svc *app.Services, _ *cobra.Command, _ []string) error {
	// Discover available tags
	tags, err := discoverAvailableTags()
	if err != nil {
		return fmt.Errorf("failed to discover available tags: %w", err)
	}
	if len(tags) < minTagsForComparison {
		return errors.New("need at least 2 tags to compare (run prof ui or prof tui to collect data first)")
	}

	// Get user selections
	selections, err := getTrackSelections(tags)
	if err != nil {
		return err
	}

	// Set global variables for the existing tracking logic
	setGlobalTrackingVariables(selections)

	// Now run the actual tracking command
	fmt.Printf("\nRunning: prof track auto --base %s --current %s --bench-name %s --profile-type %s --output-format %s",
		selections.Baseline, selections.Current, selections.BenchmarkName, selections.ProfileType, selections.OutputFormat)
	if selections.UseThreshold {
		fmt.Printf(" --fail-on-regression --regression-threshold %.1f", selections.RegressionThreshold)
	}
	fmt.Println()

	return svc.Tracker.RunTrackAuto(selections)
}

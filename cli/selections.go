package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/AlecAivazis/survey/v2"

	"github.com/AlexsanderHamir/prof/internal/app"
)

// getTrackSelections collects all user selections interactively
func getTrackSelections(tags []string) (*app.TrackOptions, error) {
	selections := &app.TrackOptions{}

	// Select baseline tag
	baselinePrompt := &survey.Select{
		Message:  "Select baseline tag (the 'before' version) [Press Enter to select]:",
		Options:  tags,
		PageSize: tuiPageSize,
	}
	if err := survey.AskOne(baselinePrompt, &selections.Baseline, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	// Select current tag (filter out baseline)
	var currentOptions []string
	for _, tag := range tags {
		if tag != selections.Baseline {
			currentOptions = append(currentOptions, tag)
		}
	}

	currentPrompt := &survey.Select{
		Message:  "Select current tag (the 'after' version) [Press Enter to select]:",
		Options:  currentOptions,
		PageSize: tuiPageSize,
	}
	if err := survey.AskOne(currentPrompt, &selections.Current, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	// Discover and select benchmark
	if err := selectBenchmark(selections); err != nil {
		return nil, err
	}

	// Discover and select profile type
	if err := selectProfileType(selections); err != nil {
		return nil, err
	}

	// Select output format
	if err := selectOutputFormat(selections); err != nil {
		return nil, err
	}

	// Ask about regression threshold
	if err := selectRegressionThreshold(selections); err != nil {
		return nil, err
	}

	return selections, nil
}

// selectBenchmark discovers and selects a benchmark
func selectBenchmark(selections *app.TrackOptions) error {
	availableBenchmarks, err := discoverAvailableBenchmarks(selections.Baseline)
	if err != nil {
		return fmt.Errorf("failed to discover benchmarks for tag %s: %w", selections.Baseline, err)
	}
	if len(availableBenchmarks) == 0 {
		return fmt.Errorf("no benchmarks found for tag %s", selections.Baseline)
	}

	benchPrompt := &survey.Select{
		Message:  "Select benchmark to compare [Press Enter to select]:",
		Options:  availableBenchmarks,
		PageSize: tuiPageSize,
	}
	return survey.AskOne(benchPrompt, &selections.BenchmarkName, survey.WithValidator(survey.Required))
}

// selectProfileType discovers and selects a profile type
func selectProfileType(selections *app.TrackOptions) error {
	availableProfiles, err := discoverAvailableProfiles(selections.Baseline, selections.BenchmarkName)
	if err != nil {
		return fmt.Errorf("failed to discover profiles for tag %s, benchmark %s: %w", selections.Baseline, selections.BenchmarkName, err)
	}
	if len(availableProfiles) == 0 {
		return fmt.Errorf("no profiles found for tag %s, benchmark %s", selections.Baseline, selections.BenchmarkName)
	}

	profilePrompt := &survey.Select{
		Message:  "Select profile type to compare [Press Enter to select]:",
		Options:  availableProfiles,
		PageSize: tuiPageSize,
	}
	return survey.AskOne(profilePrompt, &selections.ProfileType, survey.WithValidator(survey.Required))
}

// selectOutputFormat selects the output format
func selectOutputFormat(selections *app.TrackOptions) error {
	outputFormats := app.TrackOutputFormats()
	formatPrompt := &survey.Select{
		Message:  "Select output format [Press Enter to select]:",
		Options:  outputFormats,
		Default:  "detailed",
		PageSize: tuiPageSize,
	}
	return survey.AskOne(formatPrompt, &selections.OutputFormat, survey.WithValidator(survey.Required))
}

// selectRegressionThreshold handles regression threshold selection
func selectRegressionThreshold(selections *app.TrackOptions) error {
	fmt.Fprintln(os.Stdout, "By default, regression limits come from Settings → Track in prof ui (prof.json).")
	fmt.Fprintln(os.Stdout, "Enable below to override with a one-run CLI threshold for this comparison only.")

	thresholdPrompt := &survey.Confirm{
		Message: "Override prof.json and fail this run if regression exceeds a threshold?",
		Default: false,
	}
	if err := survey.AskOne(thresholdPrompt, &selections.UseThreshold); err != nil {
		return err
	}

	if selections.UseThreshold {
		var thresholdStr string
		thresholdInputPrompt := &survey.Input{
			Message: "Enter regression threshold percentage (e.g., 5.0 for 5%):",
			Default: "5.0",
		}
		if err := survey.AskOne(thresholdInputPrompt, &thresholdStr, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		threshold, convErr := strconv.ParseFloat(thresholdStr, 64)
		if convErr != nil || threshold <= 0 {
			return fmt.Errorf("invalid threshold: %s", thresholdStr)
		}
		selections.RegressionThreshold = threshold
	}

	return nil
}

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/engine/benchmark"
	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/spf13/cobra"
)

const (
	tuiPageSize          = 20
	minTagsForComparison = 2
)

func createTuiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive selection of benchmarks and profiles, then runs prof auto",
		RunE:  runTUI,
	}

	cmd.AddCommand(createTuiTrackAutoCmd())

	return cmd
}

func createTuiTrackAutoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Interactive tracking with existing benchmark data",
		RunE:  runTUITrackAuto,
	}

	return cmd
}

// discoverAvailableTags scans the bench directory for existing tags
func discoverAvailableTags() ([]string, error) {
	root, err := internal.FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root: %w", err)
	}

	benchDir := filepath.Join(root, internal.MainDirOutput)
	entries, err := os.ReadDir(benchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read bench directory: %w", err)
	}

	var tags []string
	for _, entry := range entries {
		if entry.IsDir() {
			tags = append(tags, entry.Name())
		}
	}

	return tags, nil
}

// discoverAvailableBenchmarks scans a specific tag directory for available benchmarks
func discoverAvailableBenchmarks(tag string) ([]string, error) {
	root, err := internal.FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root: %w", err)
	}

	benchDir := filepath.Join(root, internal.MainDirOutput, tag, internal.ProfileTextDir)
	entries, err := os.ReadDir(benchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read benchmark directory for tag %s: %w", tag, err)
	}

	var availableBenchmarks []string
	for _, entry := range entries {
		if entry.IsDir() {
			availableBenchmarks = append(availableBenchmarks, entry.Name())
		}
	}

	return availableBenchmarks, nil
}

// discoverAvailableProfiles scans a specific tag and benchmark for available profile types
func discoverAvailableProfiles(tag, benchmarkName string) ([]string, error) {
	root, err := internal.FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root: %w", err)
	}

	benchDir := filepath.Join(root, internal.MainDirOutput, tag, internal.ProfileTextDir, benchmarkName)
	entries, err := os.ReadDir(benchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read profile directory for tag %s, benchmark %s: %w", tag, benchmarkName, err)
	}

	var availableProfiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			// Extract profile type from filename like "BenchmarkName_cpu.txt"
			name := entry.Name()
			if strings.HasPrefix(name, benchmarkName+"_") {
				profileTypeName := strings.TrimSuffix(strings.TrimPrefix(name, benchmarkName+"_"), ".txt")
				if profileTypeName == "cpu" || profileTypeName == "memory" || profileTypeName == "mutex" || profileTypeName == "block" {
					availableProfiles = append(availableProfiles, profileTypeName)
				}
			}
		}
	}

	return availableProfiles, nil
}

func runTUI(_ *cobra.Command, _ []string) error {
	benchNames, err := benchmark.DiscoverBenchmarks()
	if err != nil {
		return fmt.Errorf("failed to discover benchmarks: %w", err)
	}

	if len(benchNames) == 0 {
		return errors.New("no benchmarks found in this module (look for func BenchmarkXxx(b *testing.B) in *_test.go)")
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

	profilesOptions := benchmark.SupportedProfiles
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

	if err = benchmark.RunBenchmarks(selectedBenches, selectedProfiles, tagStr, runCount); err != nil {
		return err
	}

	return nil
}

func runTUITrackAuto(_ *cobra.Command, _ []string) error {
	// Discover available tags
	tags, err := discoverAvailableTags()
	if err != nil {
		return fmt.Errorf("failed to discover available tags: %w", err)
	}
	if len(tags) < minTagsForComparison {
		return errors.New("need at least 2 tags to compare (run 'prof tui' first to collect some data)")
	}

	// Get user selections
	selections, err := getTrackAutoSelections(tags)
	if err != nil {
		return err
	}

	// Set global variables for the existing tracking logic
	setGlobalTrackingVariables(selections)

	// Now run the actual tracking command
	fmt.Printf("\n🚀 Running: prof track auto --base %s --current %s --bench-name %s --profile-type %s --output-format %s",
		selections.BaselineTag, selections.CurrentTag, selections.BenchmarkName, selections.ProfileType, selections.OutputFormat)
	if selections.UseThreshold {
		fmt.Printf(" --fail-on-regression --regression-threshold %.1f", selections.RegressionThreshold)
	}
	fmt.Println()

	return tracker.RunTrackAuto(selections)
}

// getTrackAutoSelections collects all user selections interactively
func getTrackAutoSelections(tags []string) (*tracker.AutoSelections, error) {
	selections := &tracker.AutoSelections{}

	// Select baseline tag
	baselinePrompt := &survey.Select{
		Message:  "Select baseline tag (the 'before' version) [Press Enter to select]:",
		Options:  tags,
		PageSize: tuiPageSize,
	}
	if err := survey.AskOne(baselinePrompt, &selections.BaselineTag, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	// Select current tag (filter out baseline)
	var currentTagOptions []string
	for _, tag := range tags {
		if tag != selections.BaselineTag {
			currentTagOptions = append(currentTagOptions, tag)
		}
	}

	currentPrompt := &survey.Select{
		Message:  "Select current tag (the 'after' version) [Press Enter to select]:",
		Options:  currentTagOptions,
		PageSize: tuiPageSize,
	}
	if err := survey.AskOne(currentPrompt, &selections.CurrentTag, survey.WithValidator(survey.Required)); err != nil {
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
func selectBenchmark(selections *tracker.AutoSelections) error {
	availableBenchmarks, err := discoverAvailableBenchmarks(selections.BaselineTag)
	if err != nil {
		return fmt.Errorf("failed to discover benchmarks for tag %s: %w", selections.BaselineTag, err)
	}
	if len(availableBenchmarks) == 0 {
		return fmt.Errorf("no benchmarks found for tag %s", selections.BaselineTag)
	}

	benchPrompt := &survey.Select{
		Message:  "Select benchmark to compare [Press Enter to select]:",
		Options:  availableBenchmarks,
		PageSize: tuiPageSize,
	}
	return survey.AskOne(benchPrompt, &selections.BenchmarkName, survey.WithValidator(survey.Required))
}

// selectProfileType discovers and selects a profile type
func selectProfileType(selections *tracker.AutoSelections) error {
	availableProfiles, err := discoverAvailableProfiles(selections.BaselineTag, selections.BenchmarkName)
	if err != nil {
		return fmt.Errorf("failed to discover profiles for tag %s, benchmark %s: %w", selections.BaselineTag, selections.BenchmarkName, err)
	}
	if len(availableProfiles) == 0 {
		return fmt.Errorf("no profiles found for tag %s, benchmark %s", selections.BaselineTag, selections.BenchmarkName)
	}

	profilePrompt := &survey.Select{
		Message:  "Select profile type to compare [Press Enter to select]:",
		Options:  availableProfiles,
		PageSize: tuiPageSize,
	}
	return survey.AskOne(profilePrompt, &selections.ProfileType, survey.WithValidator(survey.Required))
}

// selectOutputFormat selects the output format
func selectOutputFormat(selections *tracker.AutoSelections) error {
	outputFormats := []string{"summary", "detailed", "summary-html", "detailed-html", "summary-json", "detailed-json"}
	formatPrompt := &survey.Select{
		Message:  "Select output format [Press Enter to select]:",
		Options:  outputFormats,
		Default:  "detailed",
		PageSize: tuiPageSize,
	}
	return survey.AskOne(formatPrompt, &selections.OutputFormat, survey.WithValidator(survey.Required))
}

// selectRegressionThreshold handles regression threshold selection
func selectRegressionThreshold(selections *tracker.AutoSelections) error {
	thresholdPrompt := &survey.Confirm{
		Message: "Do you want to fail on performance regressions?",
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

// setGlobalTrackingVariables sets the global CLI variables for tracking
func setGlobalTrackingVariables(selections *tracker.AutoSelections) {
	baselineTag = selections.BaselineTag
	currentTag = selections.CurrentTag
	benchmarkName = selections.BenchmarkName
	profileType = selections.ProfileType
	outputFormat = selections.OutputFormat
	failOnRegression = selections.UseThreshold
	regressionThreshold = selections.RegressionThreshold
}

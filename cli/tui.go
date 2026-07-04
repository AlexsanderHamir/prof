package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/intent"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func runTUI(svc *app.Services, _ *cobra.Command, _ []string) error {
	// Get current working directory for scope-aware benchmark discovery
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	benchNames, err := svc.Collect.DiscoverBenchmarks(currentDir)
	if err != nil {
		return fmt.Errorf("failed to discover benchmarks: %w", err)
	}

	if len(benchNames) == 0 {
		return errors.New("no benchmarks found in this directory or its subdirectories (look for func BenchmarkXxx(b *testing.B) in *_test.go)")
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		termui.PrintSection(os.Stdout, int(os.Stdout.Fd()), termui.SurveySectionTitle)
	}
	surveyGap := func() {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			termui.StepGap(os.Stdout)
		}
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
	surveyGap()

	printCollectionFilterPreview(svc, selectedBenches)
	surveyGap()

	profilesOptions := svc.Collect.SupportedProfiles()
	var selectedProfiles []string
	profilesPrompt := &survey.MultiSelect{
		Message: "Select profiles:",
		Options: profilesOptions,
		Default: []string{"cpu"},
	}

	if err = survey.AskOne(profilesPrompt, &selectedProfiles, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	surveyGap()

	var countStr string
	if err = survey.AskOne(&cleanInput{Input: survey.Input{
		Message: "Number of runs (count):",
		Default: "1",
	}}, &countStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	runCount, convErr := strconv.Atoi(countStr)
	if convErr != nil || runCount < 1 {
		return fmt.Errorf("invalid count: %s", countStr)
	}
	surveyGap()

	var tagStr string
	if err = survey.AskOne(&cleanInput{Input: survey.Input{
		Message: "Tag name (used to group results under bench/<tag>):",
	}}, &tagStr, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		termui.EndSection(os.Stdout)
	}

	collect := &intent.CollectIntent{
		Benchmarks: selectedBenches,
		Profiles:   selectedProfiles,
		Tag:        tagStr,
		Count:      runCount,
	}
	collect.Normalize()
	return intent.RunValidated(collect, svc)
}

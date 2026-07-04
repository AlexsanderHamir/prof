package cli

import (
	"bufio"
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

	surveyTTY := term.IsTerminal(int(os.Stdout.Fd()))
	if surveyTTY {
		termui.PrintSection(os.Stdout, int(os.Stdout.Fd()), termui.SurveySectionTitle)
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

	printCollectionFilterPreview(os.Stdout, surveyTTY, svc, selectedBenches)

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

	reader := bufio.NewReader(os.Stdin)
	countStr, err := askConfigureLine(reader, os.Stdout, "Number of runs (count):", "1")
	if err != nil {
		return err
	}
	runCount, convErr := strconv.Atoi(countStr)
	if convErr != nil || runCount < 1 {
		return fmt.Errorf("invalid count: %s", countStr)
	}

	tagStr, err := askConfigureLine(reader, os.Stdout, "Tag name (used to group results under bench/<tag>):", "")
	if err != nil {
		return err
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

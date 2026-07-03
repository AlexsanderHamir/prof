package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/intent"
	"github.com/spf13/cobra"
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

	var selectedBenches []string
	benchPrompt := &survey.MultiSelect{
		Message:  "Select benchmarks to run:",
		Options:  benchNames,
		PageSize: tuiPageSize,
	}
	if err = survey.AskOne(benchPrompt, &selectedBenches, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	printCollectionFilterPreview(svc, selectedBenches)

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

	var groupPkg, lenient, skipPng bool
	groupPkg, lenient, skipPng, err = askAdvancedCollectOptions()
	if err != nil {
		return err
	}
	if !skipPng && !tooling.GraphvizAvailable() {
		fmt.Fprintln(os.Stdout, tooling.SkipPNGNotice)
		skipPng = true
	}

	collect := &intent.CollectIntent{
		Benchmarks:      selectedBenches,
		Profiles:        selectedProfiles,
		Tag:             tagStr,
		Count:           runCount,
		GroupByPackage:  groupPkg,
		LenientProfiles: lenient,
		SkipPNG:         skipPng,
	}
	collect.Normalize()
	return intent.RunValidated(collect, svc)
}

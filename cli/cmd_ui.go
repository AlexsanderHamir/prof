package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const profDocumentationURL = "https://alexsanderhamir.github.io/prof/"

func newUICmd(svc *app.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Interactive menu: collect, compare, tools, or setup without memorizing subcommands.",
		Long: `Open a guided menu for the most common prof workflows.

Use this when you prefer prompts to typing flags. For scripts and CI, use prof auto, prof track, and other subcommands directly.

Documentation: ` + profDocumentationURL,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUILauncher(svc)
		},
	}
}

func requireInteractiveTerminal() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("prof ui requires an interactive terminal (stdin and stdout must be TTYs). For non-interactive use, run: prof auto, prof track, prof tui, or prof -h")
	}
	return nil
}

func runUILauncher(svc *app.Services) error {
	if err := requireInteractiveTerminal(); err != nil {
		return err
	}

	choice, err := tui.RunMainMenu()
	if err != nil {
		return err
	}

	switch choice {
	case tui.MainCollect:
		return runTUI(svc, nil, nil)
	case tui.MainCompare:
		return runTUITrackAuto(svc, nil, nil)
	case tui.MainTools:
		return runUIToolsMenu(svc)
	case tui.MainSetup:
		return runUISetupWizard(svc)
	case tui.MainDocs:
		fmt.Fprintf(os.Stdout, "Prof documentation:\n  %s\n", profDocumentationURL)
		return nil
	case tui.MainQuit, tui.MainNone:
		return nil
	default:
		return fmt.Errorf("unknown action: %d", choice)
	}
}

func runUISetupWizard(svc *app.Services) error {
	confirm := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "This writes the prof configuration template into your module (same as prof setup). Continue?",
		Default: true,
	}, &confirm); err != nil {
		return err
	}
	if !confirm {
		return nil
	}
	return svc.Setup.CreateTemplate()
}

func runUIToolsMenu(svc *app.Services) error {
	const (
		toolBenchstat    = "benchstat — compare benchmark text between two tags"
		toolQcachegrind  = "qcachegrind — open a binary profile in qcachegrind"
		toolBack         = "Back to main menu"
	)
	var tool string
	if err := survey.AskOne(&survey.Select{
		Message:  "Which tool?",
		Options:  []string{toolBenchstat, toolQcachegrind, toolBack},
		PageSize: 4,
	}, &tool, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	switch tool {
	case toolBenchstat:
		return runUIBenchstat(svc)
	case toolQcachegrind:
		return runUIQcachegrind(svc)
	case toolBack:
		return runUILauncher(svc)
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}
}

func runUIBenchstat(svc *app.Services) error {
	tags, err := discoverAvailableTags()
	if err != nil {
		return fmt.Errorf("discover tags: %w", err)
	}
	if len(tags) < 2 {
		return errors.New("need at least two tags under bench/ to run benchstat (collect data with prof auto or prof tui first)")
	}

	var base string
	if err := survey.AskOne(&survey.Select{
		Message:  "Baseline tag:",
		Options:  tags,
		PageSize: tuiPageSize,
	}, &base, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	var currentOpts []string
	for _, t := range tags {
		if t != base {
			currentOpts = append(currentOpts, t)
		}
	}
	var cur string
	if err := survey.AskOne(&survey.Select{
		Message:  "Current tag:",
		Options:  currentOpts,
		PageSize: tuiPageSize,
	}, &cur, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	benches, err := discoverAvailableBenchmarks(base)
	if err != nil {
		return err
	}
	if len(benches) == 0 {
		return fmt.Errorf("no benchmarks found under bench/%s for benchstat", base)
	}

	var bench string
	if err := survey.AskOne(&survey.Select{
		Message:  "Benchmark name:",
		Options:  benches,
		PageSize: tuiPageSize,
	}, &bench, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	return svc.Tools.RunBenchStats(base, cur, bench)
}

func runUIQcachegrind(svc *app.Services) error {
	tags, err := discoverAvailableTags()
	if err != nil {
		return fmt.Errorf("discover tags: %w", err)
	}
	if len(tags) == 0 {
		return errors.New("no tags under bench/ — collect profiles first (prof auto or prof tui)")
	}

	var tagChoice string
	if err := survey.AskOne(&survey.Select{
		Message:  "Tag:",
		Options:  tags,
		PageSize: tuiPageSize,
	}, &tagChoice, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	benches, err := discoverAvailableBenchmarks(tagChoice)
	if err != nil {
		return err
	}
	if len(benches) == 0 {
		return fmt.Errorf("no benchmarks found under bench/%s", tagChoice)
	}

	var bench string
	if err := survey.AskOne(&survey.Select{
		Message:  "Benchmark name:",
		Options:  benches,
		PageSize: tuiPageSize,
	}, &bench, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	profiles, err := discoverAvailableProfiles(tagChoice, bench)
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return fmt.Errorf("no profile types found for %s / %s", tagChoice, bench)
	}

	var profName string
	if err := survey.AskOne(&survey.Select{
		Message:  "Profile type:",
		Options:  profiles,
		PageSize: tuiPageSize,
	}, &profName, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	return svc.Tools.RunQcacheGrind(tagChoice, bench, profName)
}

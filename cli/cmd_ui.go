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
		Short: "Interactive menu: collect profiles or create prof.json without memorizing subcommands.",
		Long: `Open a guided menu for the most common prof workflows.

Use this when you prefer prompts to typing flags. For scripts and automation, use prof auto and other subcommands directly.

Documentation: ` + profDocumentationURL,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUILauncher(svc)
		},
	}
}

func requireInteractiveTerminal() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return errors.New("prof ui requires an interactive terminal (stdin and stdout must be TTYs). For non-interactive use, run: prof auto, prof tui, or prof -h")
	}
	return nil
}

func runUILauncher(svc *app.Services) error {
	if err := requireInteractiveTerminal(); err != nil {
		return err
	}

	for {
		if err := runUILauncherOnce(svc); err != nil {
			if errors.Is(err, errUILoopExit) {
				return nil
			}
			return err
		}
	}
}

func runUILauncherOnce(svc *app.Services) error {
	choice, err := tui.RunMainMenu()
	if err != nil {
		return err
	}

	var runErr error
	switch choice {
	case tui.MainCollect:
		runErr = runTUI(svc, nil, nil)
	case tui.MainConfig, tui.MainSetup:
		runErr = runUIConfigCreate(svc)
	case tui.MainQuit, tui.MainNone:
		return errUILoopExit
	default:
		return fmt.Errorf("unknown action: %d", choice)
	}

	return finishUIWorkflow(runErr)
}

func finishUIWorkflow(runErr error) error {
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", runErr)
	}
	if err := promptReturnToHub(); err != nil {
		if runErr != nil && errors.Is(err, errUILoopExit) {
			return runErr
		}
		return err
	}
	return nil
}

var errUILoopExit = errors.New("ui: exit hub loop")

func promptReturnToHub() error {
	var again bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Return to main menu?",
		Default: true,
	}, &again); err != nil {
		return err
	}
	if again {
		return nil
	}
	return errUILoopExit
}

package cli

import (
	"fmt"
	"os"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/intent"
	"github.com/spf13/cobra"
)

func newConfigCmd(svc *app.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Create or inspect prof.json beside go.mod.",
	}
	cmd.AddCommand(newConfigInitCmd(svc))
	cmd.AddCommand(newConfigPathCmd(svc))
	cmd.AddCommand(newConfigValidateCmd(svc))
	return cmd
}

func newConfigInitCmd(svc *app.Services) *cobra.Command {
	return &cobra.Command{
		Use:    "init",
		Short:  "Create default prof.json beside go.mod.",
		Hidden: false,
		RunE: func(_ *cobra.Command, _ []string) error {
			return intent.RunValidated(&intent.ConfigCreateIntent{}, svc)
		},
	}
}

func newConfigPathCmd(svc *app.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the resolved prof.json path.",
		RunE: func(_ *cobra.Command, _ []string) error {
			path, err := svc.Config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, path)
			return nil
		},
	}
}

func newConfigValidateCmd(svc *app.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Load and validate prof.json; exit non-zero on error.",
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := svc.Config.Load()
			return err
		},
	}
}

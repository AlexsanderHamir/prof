package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

func newSetupCmd(svc *app.Services) *cobra.Command {
	return &cobra.Command{
		Use:                   "setup",
		Short:                 "Generates the template configuration file.",
		DisableFlagsInUseLine: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return svc.Setup.CreateTemplate()
		},
	}
}

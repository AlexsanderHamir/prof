package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/intent"
	"github.com/spf13/cobra"
)

func newSetupCmd(svc *app.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "setup",
		Short:                 "Creates prof.json (alias for prof config init).",
		Hidden:                true,
		DisableFlagsInUseLine: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return intent.RunValidated(&intent.ConfigCreateIntent{}, svc)
		},
	}
	return cmd
}

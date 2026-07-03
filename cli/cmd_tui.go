package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

func newTuiCmd(svc *app.Services) *cobra.Command {
	return &cobra.Command{
		Use:     "tui",
		Short:   "Interactive selection of benchmarks and profiles, then runs prof auto",
		Example: "prof tui",
		RunE:    func(c *cobra.Command, args []string) error { return runTUI(svc, c, args) },
	}
}

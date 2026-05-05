package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

func newTuiCmd(svc *app.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tui",
		Short:   "Interactive selection of benchmarks and profiles, then runs prof auto",
		Example: "prof tui\nprof tui track",
		RunE:    func(c *cobra.Command, args []string) error { return runTUI(svc, c, args) },
	}
	cmd.AddCommand(newTuiTrackCmd(svc))
	return cmd
}

func newTuiTrackCmd(svc *app.Services) *cobra.Command {
	return &cobra.Command{
		Use:   "track",
		Short: "Interactive tracking with existing benchmark data",
		RunE:  func(c *cobra.Command, args []string) error { return runTUITrackAuto(svc, c, args) },
	}
}

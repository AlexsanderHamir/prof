package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

// CreateRootCmd builds the root cobra command tree.
// services may be nil; nil fields are filled via [app.Services.WithDefaults].
func CreateRootCmd(services *app.Services) *cobra.Command {
	svc := services.WithDefaults()

	root := &cobra.Command{
		Use:   "prof",
		Short: "CLI tool for organizing pprof generated data, and analyzing performance differences at the profile level.",
	}

	root.AddCommand(newManualCollectCmd(svc))
	root.AddCommand(newAutoBenchmarkCmd(svc))
	root.AddCommand(newTuiCmd(svc))
	root.AddCommand(newSetupCmd(svc))
	root.AddCommand(newTrackCmd(svc))
	root.AddCommand(newToolsCmd(svc))

	return root
}

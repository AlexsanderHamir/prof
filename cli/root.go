package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

// Version is the prof CLI version. Release builds set it via
// -ldflags=-X=github.com/AlexsanderHamir/prof/cli.Version=...
// The default is used for go install from HEAD or local go build.
var Version = "devel"

// CreateRootCmd builds the root cobra command tree.
// services may be nil; nil fields are filled via [app.Services.WithDefaults].
func CreateRootCmd(services *app.Services) *cobra.Command {
	svc := services.WithDefaults()

	root := &cobra.Command{
		Use:     "prof",
		Short:   "CLI tool for organizing pprof generated data, and analyzing performance differences at the profile level.",
		Version: Version,
	}

	root.AddCommand(newManualCollectCmd(svc))
	root.AddCommand(newAutoBenchmarkCmd(svc))
	root.AddCommand(newTuiCmd(svc))
	root.AddCommand(newSetupCmd(svc))
	root.AddCommand(newTrackCmd(svc))
	root.AddCommand(newToolsCmd(svc))

	return root
}

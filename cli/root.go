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
		Use:   "prof",
		Short: "Go benchmark profiling: collect pprof-backed runs under bench/<tag>/ and compare them.",
		Long: `Prof wraps go test and pprof so you can capture CPU, memory, mutex, and block profiles in one
workflow, store artifacts in a predictable bench/<tag>/ tree, and diff two runs when you change code.

Start interactively (no flags to memorize):

  prof ui

For automation and CI, use prof auto, prof track, and other subcommands — see prof -h and prof <command> -h.

Documentation: https://alexsanderhamir.github.io/prof/`,
		Example: `  # Guided menu (recommended first run)
  prof ui

  # Collect profiles (non-interactive)
  prof auto --benchmarks "BenchmarkFoo" --profiles "cpu,memory" --count 5 --tag baseline

  # Compare two tags
  prof track auto --base baseline --current candidate --profile-type cpu --bench-name BenchmarkFoo --output-format summary`,
		Version: Version,
	}

	root.AddCommand(newUICmd(svc))
	root.AddCommand(newManualCollectCmd(svc))
	root.AddCommand(newAutoBenchmarkCmd(svc))
	root.AddCommand(newTuiCmd(svc))
	root.AddCommand(newConfigCmd(svc))
	root.AddCommand(newSetupCmd(svc))
	root.AddCommand(newTrackCmd(svc))

	return root
}

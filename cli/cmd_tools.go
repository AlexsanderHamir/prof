package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/spf13/cobra"
)

func newToolsCmd(svc *app.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Offers many tools that can easily operate on the collected data.",
	}
	cmd.AddCommand(newBenchStatCmd(svc))
	cmd.AddCommand(newQCacheGrindCmd(svc))
	return cmd
}

func newQCacheGrindCmd(svc *app.Services) *cobra.Command {
	f := &toolsGlobal
	profilesFlag := "profiles"
	var profiles []string
	cmd := &cobra.Command{
		Use:     workspace.ToolNameQcachegrind,
		Short:   "Generate callgrind output and open qcachegrind for a collected binary profile.",
		Example: "prof tools qcachegrind --tag current --profiles cpu --bench-name BenchmarkGenPool",
		RunE: func(_ *cobra.Command, _ []string) error {
			return svc.Tools.RunQcacheGrind(f.tag, f.benchmarkName, profiles[0])
		},
	}
	cmd.Flags().StringVar(&f.benchmarkName, benchNameFlag, "", "Name of the benchmark")
	cmd.Flags().StringSliceVar(&profiles, profilesFlag, []string{}, `Profiles to use (e.g., "cpu")`)
	cmd.Flags().StringVar(&f.tag, tagFlag, "", "The tag is used to organize the results")
	_ = cmd.MarkFlagRequired(benchNameFlag)
	_ = cmd.MarkFlagRequired(profilesFlag)
	_ = cmd.MarkFlagRequired(tagFlag)
	return cmd
}

func newBenchStatCmd(svc *app.Services) *cobra.Command {
	f := &toolsGlobal
	cmd := &cobra.Command{
		Use:     workspace.ToolNameBenchstat,
		Short:   "runs benchstat on txt collected data.",
		Example: "prof tools benchstat --base baseline --current current --bench-name BenchmarkGenPool",
		RunE: func(_ *cobra.Command, _ []string) error {
			return svc.Tools.RunBenchStats(f.baseline, f.current, f.benchmarkName)
		},
	}
	cmd.Flags().StringVar(&f.baseline, baseTagFlag, "", "Name of the baseline tag")
	cmd.Flags().StringVar(&f.current, currentTagFlag, "", "Name of the current tag")
	cmd.Flags().StringVar(&f.benchmarkName, benchNameFlag, "", "Name of the benchmark")
	_ = cmd.MarkFlagRequired(baseTagFlag)
	_ = cmd.MarkFlagRequired(currentTagFlag)
	_ = cmd.MarkFlagRequired(benchNameFlag)
	return cmd
}

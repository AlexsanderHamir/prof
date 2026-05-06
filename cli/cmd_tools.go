package cli

import (
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/internal/app"
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
	profilesFlag := "profiles"
	cmd := &cobra.Command{
		Use:     internal.ToolNameQcachegrind,
		Short:   "Generate callgrind output and open qcachegrind for a collected binary profile.",
		Example: "prof  tools qcachegrind --tag `current` --profiles `cpu` --bench-name `BenchmarkGenPool`",
		RunE: func(_ *cobra.Command, _ []string) error {
			return svc.Tools.RunQcacheGrind(tag, benchmarkName, profiles[0])
		},
	}
	cmd.Flags().StringVar(&benchmarkName, benchNameFlag, "", "Name of the benchmark")
	cmd.Flags().StringSliceVar(&profiles, profilesFlag, []string{}, `Profiles to use (e.g., "cpu,memory,mutex")`)
	cmd.Flags().StringVar(&tag, tagFlag, "", "The tag is used to organize the results")
	_ = cmd.MarkFlagRequired(benchNameFlag)
	_ = cmd.MarkFlagRequired(profilesFlag)
	_ = cmd.MarkFlagRequired(tagFlag)
	return cmd
}

func newBenchStatCmd(svc *app.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:     internal.ToolNameBenchstat,
		Short:   "runs benchstat on txt collected data.",
		Example: "prof tools benchstat --base `baseline` --current `current` --bench-name `BenchmarkGenPool`",
		RunE: func(_ *cobra.Command, _ []string) error {
			return svc.Tools.RunBenchStats(Baseline, Current, benchmarkName)
		},
	}
	cmd.Flags().StringVar(&Baseline, baseTagFlag, "", "Name of the baseline tag")
	cmd.Flags().StringVar(&Current, currentTagFlag, "", "Name of the current tag")
	cmd.Flags().StringVar(&benchmarkName, benchNameFlag, "", "Name of the benchmark")
	_ = cmd.MarkFlagRequired(baseTagFlag)
	_ = cmd.MarkFlagRequired(currentTagFlag)
	_ = cmd.MarkFlagRequired(benchNameFlag)
	return cmd
}

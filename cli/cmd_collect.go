package cli

import (
	"fmt"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

type manualCollectFlags struct {
	tag string
}

type autoCollectFlags struct {
	benchmarks []string
	profiles   []string
	tag        string
	count      int
}

func newManualCollectCmd(svc *app.Services) *cobra.Command {
	f := &manualCollectFlags{}
	cmd := &cobra.Command{
		Use:     CmdManual,
		Short:   "Ingest existing pprof profile binaries and organize them under bench/<tag>/ (does not run go test).",
		Args:    cobra.MinimumNArgs(1),
		Example: fmt.Sprintf("prof %s --tag tagName cpu.prof memory.prof block.prof mutex.prof", CmdManual),
		RunE: func(_ *cobra.Command, args []string) error {
			return svc.Collect.RunManual(app.CollectManualOptions{
				Files: args,
				Tag:   f.tag,
			})
		},
	}
	cmd.Flags().StringVar(&f.tag, tagFlag, "", "The tag is used to organize the results")
	_ = cmd.MarkFlagRequired(tagFlag)
	return cmd
}

func newAutoBenchmarkCmd(svc *app.Services) *cobra.Command {
	f := &autoCollectFlags{}
	benchFlag := "benchmarks"
	profileFlag := "profiles"
	countFlag := "count"
	example := fmt.Sprintf(`prof %s --%s "BenchmarkGenPool" --%s "cpu,memory" --%s 10 --%s "tag1"`,
		CmdAuto, benchFlag, profileFlag, countFlag, tagFlag)

	cmd := &cobra.Command{
		Use:     CmdAuto,
		Short:   "Wraps `go test` and `pprof` to benchmark code and gather profiling data for performance investigations.",
		Example: example,
		RunE: func(_ *cobra.Command, _ []string) error {
			return svc.Collect.RunAuto(app.CollectAutoOptions{
				Benchmarks: f.benchmarks,
				Profiles:   f.profiles,
				Tag:        f.tag,
				Count:      f.count,
			})
		},
	}
	cmd.Flags().StringSliceVar(&f.benchmarks, benchFlag, []string{}, `Benchmarks to run (e.g., "BenchmarkGenPool")"`)
	cmd.Flags().StringSliceVar(&f.profiles, profileFlag, []string{}, `Profiles to use (e.g., "cpu,memory,mutex")`)
	cmd.Flags().StringVar(&f.tag, tagFlag, "", "The tag is used to organize the results")
	cmd.Flags().IntVar(&f.count, countFlag, 0, "Number of runs")
	_ = cmd.MarkFlagRequired(benchFlag)
	_ = cmd.MarkFlagRequired(profileFlag)
	_ = cmd.MarkFlagRequired(tagFlag)
	_ = cmd.MarkFlagRequired(countFlag)
	return cmd
}

package cli

import (
	"fmt"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/spf13/cobra"
)

func newManualCollectCmd(svc *app.Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:     internal.MANUALCMD,
		Short:   "Receives profile files and performs data collection and organization. (doesn't wrap go test)",
		Args:    cobra.MinimumNArgs(1),
		Example: fmt.Sprintf("prof %s --tag tagName cpu.prof memory.prof block.prof mutex.prof", internal.MANUALCMD),
		RunE: func(_ *cobra.Command, args []string) error {
			return svc.Collector.RunCollector(args, tag, groupByPackage)
		},
	}
	cmd.Flags().StringVar(&tag, tagFlag, "", "The tag is used to organize the results")
	cmd.Flags().BoolVar(&groupByPackage, "group-by-package", false, "Group profile data by package/module and save as organized text file")
	_ = cmd.MarkFlagRequired(tagFlag)
	return cmd
}

func newAutoBenchmarkCmd(svc *app.Services) *cobra.Command {
	benchFlag := "benchmarks"
	profileFlag := "profiles"
	countFlag := "count"
	example := fmt.Sprintf(`prof %s --%s "BenchmarkGenPool" --%s "cpu,memory" --%s 10 --%s "tag1"`,
		internal.AUTOCMD, benchFlag, profileFlag, countFlag, tagFlag)

	cmd := &cobra.Command{
		Use:     internal.AUTOCMD,
		Short:   "Wraps `go test` and `pprof` to benchmark code and gather profiling data for performance investigations.",
		Example: example,
		RunE: func(_ *cobra.Command, _ []string) error {
			return svc.Benchmark.RunBenchmarks(benchmarks, profiles, tag, count, groupByPackage, lenientProfiles, skipPNG)
		},
	}
	cmd.Flags().StringSliceVar(&benchmarks, benchFlag, []string{}, `Benchmarks to run (e.g., "BenchmarkGenPool")"`)
	cmd.Flags().StringSliceVar(&profiles, profileFlag, []string{}, `Profiles to use (e.g., "cpu,memory,mutex")`)
	cmd.Flags().StringVar(&tag, tagFlag, "", "The tag is used to organize the results")
	cmd.Flags().IntVar(&count, countFlag, 0, "Number of runs")
	cmd.Flags().BoolVar(&groupByPackage, "group-by-package", false, "Group profile data by package/module and save as organized text file")
	cmd.Flags().BoolVar(&lenientProfiles, "lenient-profiles", false, "If a profile binary is missing after bench, skip it instead of failing (later steps run only for profiles present on disk)")
	cmd.Flags().BoolVar(&skipPNG, "skip-png", false, "Allow the run to succeed even when PNG generation fails (e.g. Graphviz dot not installed); default fails on PNG errors")
	_ = cmd.MarkFlagRequired(benchFlag)
	_ = cmd.MarkFlagRequired(profileFlag)
	_ = cmd.MarkFlagRequired(tagFlag)
	_ = cmd.MarkFlagRequired(countFlag)
	return cmd
}

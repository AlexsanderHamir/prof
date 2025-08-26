package cli

import (
	"fmt"

	"github.com/AlexsanderHamir/prof/engine/benchmark"
	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/engine/tools/benchstats"
	"github.com/AlexsanderHamir/prof/engine/tools/qcachegrind"
	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/spf13/cobra"
)

// CreateRootCmd creates and returns the root cobra command.
func CreateRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "prof",
		Short: "CLI tool for organizing pprof generated data, and analyzing performance differences at the profile level.",
	}

	rootCmd.AddCommand(createProfManual())
	rootCmd.AddCommand(createProfAuto())
	rootCmd.AddCommand(createTuiCmd())
	rootCmd.AddCommand(createSetupCmd())
	rootCmd.AddCommand(createTrackCmd())
	rootCmd.AddCommand(createToolsCmd())

	return rootCmd
}

func createToolsCmd() *cobra.Command {
	shortExplanation := "Offers many tools that can easily operate on the collected data."
	cmd := &cobra.Command{
		Use:   "tools",
		Short: shortExplanation,
	}

	cmd.AddCommand(createBenchStatCmd())
	cmd.AddCommand(createQCacheGrindCmd())

	return cmd
}

func createQCacheGrindCmd() *cobra.Command {
	profilesFlag := "profiles"
	shortExplanation := "runs benchstat on txt collected data."

	cmd := &cobra.Command{
		Use:     "qcachegrind",
		Short:   shortExplanation,
		Example: "prof  tools qcachegrind --tag `current` --profiles `cpu` --bench-name `BenchmarkGenPool`",
		RunE: func(_ *cobra.Command, _ []string) error {
			return qcachegrind.RunQcacheGrind(tag, benchmarkName, profiles[0])
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

func createBenchStatCmd() *cobra.Command {
	shortExplanation := "runs benchstat on txt collected data."

	cmd := &cobra.Command{
		Use:     "benchstat",
		Short:   shortExplanation,
		Example: "prof tools benchstat --base `baseline` --current `current` --bench-name `BenchmarkGenPool`",
		RunE: func(_ *cobra.Command, _ []string) error {
			return benchstats.RunBenchStats(Baseline, Current, benchmarkName)
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

func createProfManual() *cobra.Command {
	manualCmd := &cobra.Command{
		Use:     internal.MANUALCMD,
		Short:   "Receives profile files and performs data collection and organization. (doesn't wrap go test)",
		Args:    cobra.MinimumNArgs(1),
		Example: fmt.Sprintf("prof %s --tag tagName cpu.prof memory.prof block.prof mutex.prof", internal.MANUALCMD),
		RunE: func(_ *cobra.Command, args []string) error {
			return collector.RunCollector(args, tag, groupByPackage)
		},
	}

	manualCmd.Flags().StringVar(&tag, tagFlag, "", "The tag is used to organize the results")
	manualCmd.Flags().BoolVar(&groupByPackage, "group-by-package", false, "Group profile data by package/module and save as organized text file")
	_ = manualCmd.MarkFlagRequired(tagFlag)

	return manualCmd
}

func createProfAuto() *cobra.Command {
	benchFlag := "benchmarks"
	profileFlag := "profiles"
	countFlag := "count"
	example := fmt.Sprintf(`prof %s --%s "BenchmarkGenPool" --%s "cpu,memory" --%s 10 --%s "tag1"`, internal.AUTOCMD, benchFlag, profileFlag, countFlag, tagFlag)

	cmd := &cobra.Command{
		Use:   internal.AUTOCMD,
		Short: "Wraps `go test` and `pprof` to benchmark code and gather profiling data for performance investigations.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return benchmark.RunBenchmarks(benchmarks, profiles, tag, count, groupByPackage)
		},
		Example: example,
	}

	cmd.Flags().StringSliceVar(&benchmarks, benchFlag, []string{}, `Benchmarks to run (e.g., "BenchmarkGenPool")"`)
	cmd.Flags().StringSliceVar(&profiles, profileFlag, []string{}, `Profiles to use (e.g., "cpu,memory,mutex")`)
	cmd.Flags().StringVar(&tag, tagFlag, "", "The tag is used to organize the results")
	cmd.Flags().IntVar(&count, countFlag, 0, "Number of runs")
	cmd.Flags().BoolVar(&groupByPackage, "group-by-package", false, "Group profile data by package/module and save as organized text file")

	_ = cmd.MarkFlagRequired(benchFlag)
	_ = cmd.MarkFlagRequired(profileFlag)
	_ = cmd.MarkFlagRequired(tagFlag)
	_ = cmd.MarkFlagRequired(countFlag)

	return cmd
}

func createTrackCmd() *cobra.Command {
	shortExplanation := "Compare performance between two benchmark runs to detect regressions and improvements"
	cmd := &cobra.Command{
		Use:   "track",
		Short: shortExplanation,
	}

	cmd.AddCommand(createTrackAutoCmd())
	cmd.AddCommand(createTrackManualCmd())

	return cmd
}

func createTrackAutoCmd() *cobra.Command {
	profileTypeFlag := "profile-type"
	outputFormatFlag := "output-format"
	failFlag := "fail-on-regression"
	thresholdFlag := "regression-threshold"
	example := fmt.Sprintf(`prof track auto --%s "tag1" --%s "tag2" --%s "cpu" --%s "BenchmarkGenPool" --%s "summary"`, baseTagFlag, currentTagFlag, profileTypeFlag, benchNameFlag, outputFormatFlag)
	longExplanation := fmt.Sprintf("This command only works if the %s command was used to collect and organize the benchmark and profile data, as it expects a specific directory structure generated by that process.", internal.AUTOCMD)
	shortExplanation := "If prof auto was used to collect the data, track auto can be used to analyze it, you just have to pass the tag name."

	cmd := &cobra.Command{
		Use:   internal.TrackAutoCMD,
		Short: shortExplanation,
		Long:  longExplanation,
		RunE: func(_ *cobra.Command, _ []string) error {
			selections := &tracker.Selections{
				OutputFormat:        outputFormat,
				Baseline:            Baseline,
				Current:             Current,
				ProfileType:         profileType,
				BenchmarkName:       benchmarkName,
				RegressionThreshold: regressionThreshold,
				UseThreshold:        failOnRegression,
			}
			return tracker.RunTrackAuto(selections)
		},
		Example: example,
	}

	cmd.Flags().StringVar(&Baseline, baseTagFlag, "", "Name of the baseline tag")
	cmd.Flags().StringVar(&Current, currentTagFlag, "", "Name of the current tag")
	cmd.Flags().StringVar(&benchmarkName, benchNameFlag, "", "Name of the benchmark")
	cmd.Flags().StringVar(&profileType, profileTypeFlag, "", "Profile type (cpu, memory, mutex, block)")
	cmd.Flags().StringVar(&outputFormat, outputFormatFlag, "detailed", `Output format: "summary" or "detailed"`)
	cmd.Flags().BoolVar(&failOnRegression, failFlag, false, "Exit with non-zero code if regression exceeds threshold (optional when using CI/CD config)")
	cmd.Flags().Float64Var(&regressionThreshold, thresholdFlag, 0.0, "Fail when worst flat regression exceeds this percent (optional when using CI/CD config)")

	_ = cmd.MarkFlagRequired(baseTagFlag)
	_ = cmd.MarkFlagRequired(currentTagFlag)
	_ = cmd.MarkFlagRequired(benchNameFlag)
	_ = cmd.MarkFlagRequired(profileTypeFlag)

	return cmd
}

func createTrackManualCmd() *cobra.Command {
	outputFormatFlag := "output-format"
	failFlag := "fail-on-regression"
	thresholdFlag := "regression-threshold"
	example := fmt.Sprintf(`prof track %s --%s "path/to/profile_file.txt" --%s  "path/to/profile_file.txt"  --%s "summary"`, internal.TrackManualCMD, baseTagFlag, currentTagFlag, outputFormatFlag)

	cmd := &cobra.Command{
		Use:   internal.TrackManualCMD,
		Short: "Manually specify the paths to the profile text files you want to compare.",
		RunE: func(_ *cobra.Command, _ []string) error {
			selections := &tracker.Selections{
				OutputFormat:        outputFormat,
				Baseline:            Baseline,
				Current:             Current,
				ProfileType:         profileType,
				BenchmarkName:       benchmarkName,
				RegressionThreshold: regressionThreshold,
				UseThreshold:        failOnRegression,
				IsManual:            true,
			}
			return tracker.RunTrackManual(selections)
		},
		Example: example,
	}

	cmd.Flags().StringVar(&Baseline, baseTagFlag, "", "Name of the baseline tag")
	cmd.Flags().StringVar(&Current, currentTagFlag, "", "Name of the current tag")
	cmd.Flags().StringVar(&outputFormat, outputFormatFlag, "", "Output format choice choice")
	cmd.Flags().BoolVar(&failOnRegression, failFlag, false, "Exit with non-zero code if regression exceeds threshold (optional when using CI/CD config)")
	cmd.Flags().Float64Var(&regressionThreshold, thresholdFlag, 0.0, "Fail when worst flat regression exceeds this percent (optional when using CI/CD config)")

	_ = cmd.MarkFlagRequired(baseTagFlag)
	_ = cmd.MarkFlagRequired(currentTagFlag)
	_ = cmd.MarkFlagRequired(outputFormatFlag)

	return cmd
}

func createSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Generates the template configuration file.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return internal.CreateTemplate()
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func createTuiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive selection of benchmarks and profiles, then runs prof auto",
		RunE:  runTUI,
	}

	cmd.AddCommand(createTuiTrackAutoCmd())

	return cmd
}

func createTuiTrackAutoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Interactive tracking with existing benchmark data",
		RunE:  runTUITrackAuto,
	}

	return cmd
}

package cli

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/tracker"
	"github.com/AlexsanderHamir/prof/version"
	"github.com/spf13/cobra"
)

var (
	// Root command flags.
	benchmarks string
	profiles   string
	tag        string
	count      int

	// Track command flags.
	baselineTag   string
	currentTag    string
	benchmarkName string
	profileType   string
	outputFormat  string
)

// CreateRootCmd creates and returns the root cobra command.
func CreateRootCmd() *cobra.Command {
	benchFlag := "benchmarks"
	profileFlag := "profiles"
	tagFlag := "tag"
	countFlag := "count"
	example := fmt.Sprintf(`prof --%s "[BenchmarkGenPool]" --%s "[cpu,memory]" --%s 10 --%s "tag1"`,
		benchFlag, profileFlag, countFlag, tagFlag)

	rootCmd := &cobra.Command{
		Use:     "prof",
		Short:   "CLI tool for organizing pprof generated data, and analyzing performance differences at the function level.",
		RunE:    runBenchmarks,
		Example: example,
	}

	rootCmd.Flags().StringVar(&benchmarks, benchFlag, "", `Benchmarks to run (e.g., "[BenchmarkGenPool]")"`)
	rootCmd.Flags().StringVar(&profiles, profileFlag, "", `Profiles to use (e.g., "[cpu,memory,mutex]")`)
	rootCmd.Flags().StringVar(&tag, tagFlag, "", "Tag for the run")
	rootCmd.Flags().IntVar(&count, countFlag, 0, "Number of runs")

	_ = rootCmd.MarkFlagRequired(benchFlag)
	_ = rootCmd.MarkFlagRequired(profileFlag)
	_ = rootCmd.MarkFlagRequired(tagFlag)
	_ = rootCmd.MarkFlagRequired(countFlag)

	rootCmd.AddCommand(createSetupCmd())
	rootCmd.AddCommand(createTrackCmd())
	rootCmd.AddCommand(createVersionCmd())

	return rootCmd
}

// createTrackCmd creates the track subcommand
func createTrackCmd() *cobra.Command {
	baseTagFlag := "base-tag"
	currentTagFlag := "current-tag"
	benchNameFlag := "bench-name"
	profileTypeFlag := "profile-type"
	outputFormatFlag := "output-format"
	example := fmt.Sprintf(`prof track --%s "tag1" --%s "tag2" --%s "cpu" --%s "BenchmarkGenPool" --%s "summary"`,
		baseTagFlag, currentTagFlag, profileTypeFlag, benchNameFlag, outputFormatFlag)

	cmd := &cobra.Command{
		Use:     "track",
		Short:   "Compare performance between two benchmark runs to detect regressions and improvements",
		RunE:    runTrack,
		Example: example,
	}

	cmd.Flags().StringVar(&baselineTag, baseTagFlag, "", "Name of the baseline tag")
	cmd.Flags().StringVar(&currentTag, currentTagFlag, "", "Name of the current tag")
	cmd.Flags().StringVar(&benchmarkName, benchNameFlag, "", "Name of the benchmark")
	cmd.Flags().StringVar(&profileType, profileTypeFlag, "", "Profile type (cpu, memory, mutex, block)")
	cmd.Flags().StringVar(&outputFormat, outputFormatFlag, "detailed", `Output format: "summary" or "detailed"`)

	_ = cmd.MarkFlagRequired(baseTagFlag)
	_ = cmd.MarkFlagRequired(currentTagFlag)
	_ = cmd.MarkFlagRequired(benchNameFlag)
	_ = cmd.MarkFlagRequired(profileTypeFlag)

	return cmd
}

func createVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "version",
		Short:                 "Shows the current version of prof and checks for updates.",
		RunE:                  runVersion,
		DisableFlagsInUseLine: true,
	}

	return cmd
}

// createSetupCmd creates the setup subcommand
func createSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "setup",
		Short:                 "Generates the template configuration file.",
		RunE:                  runSetup,
		DisableFlagsInUseLine: true,
	}

	return cmd
}

// Execute runs the CLI application
func Execute() error {
	return CreateRootCmd().Execute()
}

func runBenchmarks(_ *cobra.Command, _ []string) error {
	if benchmarks == "" || profiles == "" || tag == "" || count == 0 {
		return errors.New("missing required arguments. Use --help for usage information")
	}

	cfg, err := config.LoadFromFile("config_template.json")
	if err != nil {
		cfg = &config.Config{}
	}

	benchmarkList, profileList, err := parseBenchmarkConfig(benchmarks, profiles)
	if err != nil {
		return fmt.Errorf("failed to parse benchmark config: %w", err)
	}

	if err = setupDirectories(tag, benchmarkList, profileList); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	benchArgs := &args.BenchArgs{
		Benchmarks: benchmarkList,
		Profiles:   profileList,
		Count:      count,
		Tag:        tag,
	}

	printConfiguration(benchArgs, cfg.FunctionFilter)

	if err = runBencAndGetProfiles(benchArgs, cfg.FunctionFilter); err != nil {
		return err
	}

	return nil
}

// runVersion handles the version command execution
func runVersion(_ *cobra.Command, _ []string) error {
	current, latest := version.Check()
	output := version.FormatOutput(current, latest)
	fmt.Print(output)
	return nil
}

// runSetup handles the setup command execution
func runSetup(_ *cobra.Command, _ []string) error {
	return config.CreateTemplate()
}

// runTrack handles the track command execution
func runTrack(_ *cobra.Command, _ []string) error {
	if !validProfiles[profileType] {
		return fmt.Errorf("invalid profile type '%s'. Valid types: cpu, memory, mutex, block", profileType)
	}

	validFormats := map[string]bool{
		"summary":  true,
		"detailed": true,
	}

	if !validFormats[outputFormat] {
		return fmt.Errorf("invalid output format '%s'. Valid formats: summary, detailed", outputFormat)
	}

	// Call the tracker API
	report, err := tracker.CheckPerformanceDifferences(
		baselineTag,
		currentTag,
		benchmarkName,
		profileType,
	)

	if err != nil {
		return fmt.Errorf("failed to track performance differences: %w", err)
	}

	// Display results based on output format
	if len(report.FunctionChanges) == 0 {
		slog.Info("No function changes detected between the two runs")
		return nil
	}

	switch outputFormat {
	case "summary":
		printSummary(report)
	case "detailed":
		printDetailedReport(report)
	}

	return nil
}

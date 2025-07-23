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

	// Setup command flags.
	createTemplate bool
	outputPath     string

	// Track command flags.
	baselineTag   string
	currentTag    string
	benchmarkName string
	profileType   string
	outputFormat  string
)

// CreateRootCmd creates and returns the root cobra command.
func CreateRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "prof",
		Short: "CLI tool for organizing and analyzing Go benchmarks with AI",
		Long: `Prof is a comprehensive tool for running Go benchmarks, collecting performance profiles,
and analyzing them with AI to identify performance bottlenecks and improvements.`,
		RunE: runBenchmarks,
	}

	benchFlag := "benchmarks"
	profileFlag := "profiles"
	tagFlag := "tag"
	countFlag := "count"

	rootCmd.Flags().StringVar(&benchmarks, benchFlag, "", "Benchmarks to run (e.g., '[BenchmarkGenPool,...]')")
	rootCmd.Flags().StringVar(&profiles, profileFlag, "", "Profiles to use (e.g., '[cpu,memory,mutex]')")
	rootCmd.Flags().StringVar(&tag, tagFlag, "", "Tag for the run")
	rootCmd.Flags().IntVar(&count, countFlag, 0, "Number of runs")

	rootCmd.MarkFlagRequired(benchFlag)   //nolint:errcheck // won't fail — flags are created just above
	rootCmd.MarkFlagRequired(profileFlag) //nolint:errcheck // won't fail — flags are created just above
	rootCmd.MarkFlagRequired(tagFlag)     //nolint:errcheck // won't fail — flags are created just above
	rootCmd.MarkFlagRequired(countFlag)   //nolint:errcheck // won't fail — flags are created just above

	// Add subcommands.
	rootCmd.AddCommand(createSetupCmd())
	rootCmd.AddCommand(createTrackCmd())
	rootCmd.AddCommand(createVersionCmd())

	return rootCmd
}

func createVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the current version of prof and check for available updates.`,
		RunE:  runVersion,
	}

	return cmd
}

// createSetupCmd creates the setup subcommand
func createSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up configuration for the benchmarking tool",
		Long:  `Generate template configuration files and set up the benchmarking environment.`,
		RunE:  runSetup,
	}

	createTemplateFlag := "create-template"
	outputPathFlagh := "output-path"

	cmd.Flags().BoolVar(&createTemplate, createTemplateFlag, false, "Generate a new template configuration file")
	cmd.Flags().StringVar(&outputPath, outputPathFlagh, "./config_template.json", "Destination path for the template")

	cmd.MarkFlagRequired(createTemplateFlag) //nolint:errcheck // won't fail — flags are created just above

	return cmd
}

// createTrackCmd creates the track subcommand
func createTrackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Compare performance between two benchmark runs to detect regressions and improvements",
		Long: `Compare performance differences between two benchmark runs identified by their tags.
This command analyzes profile data to detect regressions and improvements, providing
detailed reports on function-level performance changes including execution time deltas
and percentage changes.

Example:
  prof track --base-tag v1.0 --current-tag v1.1 --bench BenchmarkPool --profile-type cpu`,
		RunE: runTrack,
	}

	cmd.Flags().StringVar(&baselineTag, "base-tag", "", "Name of the baseline tag")
	cmd.Flags().StringVar(&currentTag, "current-tag", "", "Name of the current tag")
	cmd.Flags().StringVar(&benchmarkName, "bench", "", "Name of the benchmark")
	cmd.Flags().StringVar(&profileType, "profile-type", "", "Profile type (cpu, memory, mutex, block)")
	cmd.Flags().StringVar(&outputFormat, "format", "detailed", "Output format: 'summary' or 'detailed'")

	// Mark required flags
	cmd.MarkFlagRequired("base-tag")     //nolint:errcheck // won't fail — flags are created just above
	cmd.MarkFlagRequired("current-tag")  //nolint:errcheck // won't fail — flags are created just above
	cmd.MarkFlagRequired("bench")        //nolint:errcheck // won't fail — flags are created just above
	cmd.MarkFlagRequired("profile-type") //nolint:errcheck // won't fail — flags are created just above

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
	if createTemplate {
		return config.CreateTemplate(outputPath)
	}
	return errors.New("setup command requires --create-template flag")
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

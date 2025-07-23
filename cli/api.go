package cli

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/shared"
	"github.com/AlexsanderHamir/prof/tracker"
	"github.com/AlexsanderHamir/prof/version"
	"github.com/spf13/cobra"
)

var (
	// Root command flags
	showVersion    bool
	benchmarks     string
	profiles       string
	tag            string
	count          int
	generalAnalyze bool
	flagProfiles   bool

	// Setup command flags
	createTemplate bool
	outputPath     string

	// Track command flags
	baselineTag   string
	currentTag    string
	benchmarkName string
	profileType   string
	outputFormat  string
)

// CreateRootCmd creates and returns the root cobra command
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

	rootCmd.Flags().StringVar(&benchmarks, benchFlag, "", "Benchmarks to run (e.g., '[BenchmarkGenPool, BenchmarkSyncPool]')")
	rootCmd.Flags().StringVar(&profiles, profileFlag, "", "Profiles to use (e.g., '[cpu,memory,mutex]')")
	rootCmd.Flags().StringVar(&tag, tagFlag, "", "Tag for the run")
	rootCmd.Flags().IntVar(&count, countFlag, 0, "Number of runs")

	rootCmd.MarkFlagRequired(benchFlag)
	rootCmd.MarkFlagRequired(profileFlag)
	rootCmd.MarkFlagRequired(tagFlag)
	rootCmd.MarkFlagRequired(countFlag)

	// TODO:
	// There's no need for AI analysis to run together with the benchmarks,
	// it can easily run afterwards as a seprate command.
	rootCmd.Flags().BoolVar(&generalAnalyze, "general-analyze", false, "Run general AI analysis")
	rootCmd.Flags().BoolVar(&flagProfiles, "flag-profiles", false, "Flag profiles for review")

	// Add subcommands
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

	cmd.MarkFlagRequired(createTemplateFlag)
	cmd.MarkFlagRequired(outputPathFlagh)

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
	cmd.MarkFlagRequired("base-tag")
	cmd.MarkFlagRequired("current-tag")
	cmd.MarkFlagRequired("bench")
	cmd.MarkFlagRequired("profile-type")

	return cmd
}

// Execute runs the CLI application
func Execute() error {
	return CreateRootCmd().Execute()
}

// runBenchmarks handles the root command execution
func runBenchmarks(_ *cobra.Command, _ []string) error {
	if showVersion {
		current, latest := version.Check()
		output := version.FormatOutput(current, latest)
		fmt.Print(output)
		return nil
	}

	// Validate required arguments for benchmark run
	if benchmarks == "" || profiles == "" || tag == "" || count == 0 {
		return fmt.Errorf("missing required arguments. Use --help for usage information")
	}

	// Load config
	cfg, err := config.LoadFromFile("config_template.json")
	if err != nil {
		cfg = &config.Config{} // use default
	}

	// Parse benchmark config
	benchmarkList, profileList, err := parseBenchmarkConfig(benchmarks, profiles)
	if err != nil {
		return fmt.Errorf("failed to parse benchmark config: %w", err)
	}

	// Setup directories
	if err := setupDirectories(tag, benchmarkList, profileList); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	benchArgs := &args.BenchArgs{
		Benchmarks: benchmarkList,
		Profiles:   profileList,
		Count:      count,
		Tag:        tag,
	}

	printConfiguration(benchArgs, cfg.FunctionFilter)

	// Run benchmarks
	if err := runBencAndGetProfiles(benchArgs, cfg.FunctionFilter); err != nil {
		return err
	}

	// Run AI analysis if requested
	if generalAnalyze {
		if err := analyzeProfiles(tag, profileList, cfg, flagProfiles); err != nil {
			return fmt.Errorf("failed to analyze profiles: %w", err)
		}
	}

	// Flag profiles if requested
	if flagProfiles {
		if err := analyzeProfiles(tag, profileList, cfg, flagProfiles); err != nil {
			return fmt.Errorf("failed to flag profiles: %w", err)
		}
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
func runSetup(cmd *cobra.Command, args []string) error {
	if createTemplate {
		return config.CreateTemplate(outputPath)
	}
	return fmt.Errorf("setup command requires --create-template flag")
}

// runTrack handles the track command execution
func runTrack(_ *cobra.Command, _ []string) error {

	if !validProfiles[profileType] {
		return fmt.Errorf("invalid profile type '%s'. Valid types: cpu, memory, mutex, block", profileType)
	}

	// Validate output format
	validFormats := map[string]bool{
		"summary":  true,
		"detailed": true,
	}
	if !validFormats[outputFormat] {
		return fmt.Errorf("invalid output format '%s'. Valid formats: summary, detailed", outputFormat)
	}

	slog.Info("Starting performance tracking",
		"baseline", baselineTag,
		"current", currentTag,
		"benchmark", benchmarkName,
		"profile", profileType)

	// Construct tag paths
	baselineTagPath := filepath.Join(shared.MainDirOutput, baselineTag)
	currentTagPath := filepath.Join(shared.MainDirOutput, currentTag)

	// Call the tracker API
	report, err := tracker.CheckPerformanceDifferences(
		baselineTagPath,
		currentTagPath,
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

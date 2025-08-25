package tracker

import (
	"fmt"
	"log/slog"
	"strings"

	"math"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

// trackAutoSelections holds all the user selections for tracking
type Selections struct {
	Baseline            string
	Current             string
	BenchmarkName       string
	ProfileType         string
	OutputFormat        string
	UseThreshold        bool
	RegressionThreshold float64
	IsManual            bool
}

var validFormats = map[string]bool{
	"summary":       true,
	"detailed":      true,
	"summary-html":  true,
	"detailed-html": true,
	"summary-json":  true,
	"detailed-json": true,
}

// runTrack handles the track command execution
func RunTrackAuto(selections *Selections) error {
	if !validFormats[selections.OutputFormat] {
		return fmt.Errorf("invalid output format '%s'. Valid formats: summary, detailed", selections.OutputFormat)
	}

	report, err := CheckPerformanceDifferences(selections)
	if err != nil {
		return fmt.Errorf("failed to track performance differences: %w", err)
	}

	noFunctionChanges := len(report.FunctionChanges) == 0
	if noFunctionChanges {
		slog.Info("No function changes detected between the two runs")
		return nil
	}

	report.ChooseOutputFormat(selections.OutputFormat)

	// Apply CI/CD filtering and thresholds
	if err = applyCIConfiguration(report, selections); err != nil {
		return err
	}

	return nil
}

// RunTrackManual receives the location of the .out / .prof files.
func RunTrackManual(selections *Selections) error {
	if !validFormats[selections.OutputFormat] {
		return fmt.Errorf("invalid output format '%s'. Valid formats: summary, detailed", selections.OutputFormat)
	}

	report, err := CheckPerformanceDifferences(selections)
	if err != nil {
		return fmt.Errorf("failed to track performance differences: %w", err)
	}

	noFunctionChanges := len(report.FunctionChanges) == 0
	if noFunctionChanges {
		slog.Info("No function changes detected between the two runs")
		return nil
	}

	report.ChooseOutputFormat(selections.OutputFormat)

	// Apply CI/CD filtering and thresholds
	if err = applyCIConfiguration(report, selections); err != nil {
		return err
	}

	return nil
}

// applyCIConfiguration applies CI/CD configuration to the performance report
func applyCIConfiguration(report *ProfileChangeReport, selections *Selections) error {
	// Load CI/CD configuration
	cfg, err := internal.LoadFromFile(internal.ConfigFilename)
	if err != nil {
		slog.Info("No CI/CD configuration found, using command-line settings only")
		// Fall back to command-line threshold logic
		return applyCommandLineThresholds(report, selections)
	}

	// Apply CI/CD filtering
	report.ApplyCIConfiguration(cfg.CIConfig, selections.BenchmarkName)

	// Check if CLI flags were provided for regression checking
	cliFlagsProvided := selections.UseThreshold || selections.RegressionThreshold > 0.0

	if cliFlagsProvided {
		// User provided CLI flags, use them (with CI/CD config as fallback)
		return applyCommandLineThresholds(report, selections)
	}

	// No CLI flags provided, use CI/CD config only
	slog.Info("No CLI regression flags provided, using CI/CD configuration settings")
	return applyCICDThresholdsOnly(report, selections, cfg.CIConfig)
}

// applyCommandLineThresholds applies the legacy command-line threshold logic
func applyCommandLineThresholds(report *ProfileChangeReport, selections *Selections) error {
	if selections.UseThreshold && selections.RegressionThreshold > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent >= selections.RegressionThreshold {
			return fmt.Errorf("performance regression %.2f%% in %s exceeds threshold %.2f%%", worst.FlatChangePercent, worst.FunctionName, selections.RegressionThreshold)
		}
	}
	return nil
}

// applyCICDThresholdsOnly applies CI/CD specific threshold logic only, without CLI flags
func applyCICDThresholdsOnly(report *ProfileChangeReport, selections *Selections, cicdConfig *internal.CIConfig) error {
	benchmarkName := selections.BenchmarkName
	if benchmarkName == "" {
		benchmarkName = "unknown"
	}

	// Get effective regression threshold
	effectiveThreshold := getEffectiveRegressionThreshold(cicdConfig, benchmarkName, 0.0) // Use 0.0 for no CLI threshold

	// Get minimum change threshold
	minChangeThreshold := getMinChangeThreshold(cicdConfig, benchmarkName)

	// Check if we should fail on improvements
	failOnImprovement := shouldFailOnImprovement(cicdConfig, benchmarkName)

	// Apply thresholds
	if effectiveThreshold > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent >= effectiveThreshold {
			// Check if function should be ignored by CI/CD config
			if !shouldIgnoreFunction(cicdConfig, worst.FunctionName, benchmarkName) {
				return fmt.Errorf("performance regression %.2f%% in %s exceeds CI/CD threshold %.2f%%",
					worst.FlatChangePercent, worst.FunctionName, effectiveThreshold)
			}
		}
	}

	// Check for improvements if configured to fail on them
	if failOnImprovement {
		best := report.BestImprovement()
		if best != nil && math.Abs(best.FlatChangePercent) >= minChangeThreshold {
			if !shouldIgnoreFunction(cicdConfig, best.FunctionName, benchmarkName) {
				return fmt.Errorf("unexpected performance improvement %.2f%% in %s (configured to fail on improvements)",
					math.Abs(best.FlatChangePercent), best.FunctionName)
			}
		}
	}

	// Check minimum change threshold
	if minChangeThreshold > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent < minChangeThreshold {
			slog.Info("Performance regression below minimum threshold, not failing CI/CD",
				"function", worst.FunctionName,
				"change", worst.FlatChangePercent,
				"threshold", minChangeThreshold)
		}
	}

	return nil
}

// Helper functions for CI/CD configuration
func getEffectiveRegressionThreshold(cicdConfig *internal.CIConfig, benchmarkName string, commandLineThreshold float64) float64 {
	if cicdConfig == nil {
		return commandLineThreshold
	}

	// Check benchmark-specific config first
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists && benchmarkConfig.MaxRegressionThreshold > 0 {
		if commandLineThreshold > 0 {
			if commandLineThreshold < benchmarkConfig.MaxRegressionThreshold {
				return commandLineThreshold
			}
			return benchmarkConfig.MaxRegressionThreshold
		}
		return benchmarkConfig.MaxRegressionThreshold
	}

	// Fall back to global config
	if cicdConfig.Global != nil && cicdConfig.Global.MaxRegressionThreshold > 0 {
		if commandLineThreshold > 0 {
			if commandLineThreshold < cicdConfig.Global.MaxRegressionThreshold {
				return commandLineThreshold
			}
			return cicdConfig.Global.MaxRegressionThreshold
		}
		return cicdConfig.Global.MaxRegressionThreshold
	}

	return commandLineThreshold
}

func getMinChangeThreshold(cicdConfig *internal.CIConfig, benchmarkName string) float64 {
	if cicdConfig == nil {
		return 0.0
	}

	// Check benchmark-specific config first
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists && benchmarkConfig.MinChangeThreshold > 0 {
		return benchmarkConfig.MinChangeThreshold
	}

	// Fall back to global config
	if cicdConfig.Global != nil && cicdConfig.Global.MinChangeThreshold > 0 {
		return cicdConfig.Global.MinChangeThreshold
	}

	return 0.0
}

func shouldFailOnImprovement(cicdConfig *internal.CIConfig, benchmarkName string) bool {
	if cicdConfig == nil {
		return false
	}

	// Check benchmark-specific config first
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists {
		return benchmarkConfig.FailOnImprovement
	}

	// Fall back to global config
	if cicdConfig.Global != nil {
		return cicdConfig.Global.FailOnImprovement
	}

	return false
}

func shouldIgnoreFunction(cicdConfig *internal.CIConfig, functionName string, benchmarkName string) bool {
	if cicdConfig == nil {
		return false
	}

	// Check benchmark-specific config first
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists {
		if shouldIgnoreFunctionByConfig(&benchmarkConfig, functionName) {
			return true
		}
	}

	// Fall back to global config
	if cicdConfig.Global != nil {
		return shouldIgnoreFunctionByConfig(cicdConfig.Global, functionName)
	}

	return false
}

func shouldIgnoreFunctionByConfig(config *internal.CITrackingConfig, functionName string) bool {
	if config == nil {
		return false
	}

	// Check exact function name matches
	for _, ignoredFunc := range config.IgnoreFunctions {
		if functionName == ignoredFunc {
			return true
		}
	}

	// Check prefix matches
	for _, ignoredPrefix := range config.IgnorePrefixes {
		if strings.HasPrefix(functionName, ignoredPrefix) {
			return true
		}
	}

	return false
}

// CheckPerformanceDifferences creates the profile report by comparing data from  prof's auto run.
func CheckPerformanceDifferences(selections *Selections) (*ProfileChangeReport, error) {
	binFilePathBaseLine, binFilePathCurrent := chooseFileLocations(selections)

	lineObjsBaseline, err := parser.TurnLinesIntoObjectsV2(binFilePathBaseLine)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", binFilePathBaseLine, err)
	}

	lineObjsCurrent, err := parser.TurnLinesIntoObjectsV2(binFilePathCurrent)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", binFilePathCurrent, err)
	}

	matchingMap := createMapFromLineObjects(lineObjsBaseline)

	pgp := &ProfileChangeReport{}
	for _, currentObj := range lineObjsCurrent {
		baseLineObj, matchNotFound := matchingMap[currentObj.FnName]
		if !matchNotFound {
			continue
		}

		var changeResult *FunctionChangeResult
		changeResult, err = detectChangeBetweenTwoObjects(baseLineObj, currentObj)
		if err != nil {
			return nil, fmt.Errorf("detectChangeBetweenTwoObjects failed: %w", err)
		}

		pgp.FunctionChanges = append(pgp.FunctionChanges, changeResult)
	}

	return pgp, nil
}

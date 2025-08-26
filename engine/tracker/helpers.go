package tracker

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func createMapFromLineObjects(lineobjects []*parser.LineObj) map[string]*parser.LineObj {
	matchingMap := make(map[string]*parser.LineObj)
	for _, lineObj := range lineobjects {
		matchingMap[lineObj.FnName] = lineObj
	}

	return matchingMap
}

func detectChangeBetweenTwoObjects(baseline, current *parser.LineObj) (*FunctionChangeResult, error) {
	if current == nil {
		return nil, errors.New("current obj is nil")
	}
	if baseline == nil {
		return nil, errors.New("baseLine obj is nil")
	}

	const percentMultiplier = 100

	var flatChange float64
	if baseline.Flat != 0 {
		flatChange = ((current.Flat - baseline.Flat) / baseline.Flat) * percentMultiplier
	}

	var cumChange float64
	if baseline.Cum != 0 {
		cumChange = ((current.Cum - baseline.Cum) / baseline.Cum) * percentMultiplier
	}

	changeType := internal.STABLE
	if flatChange > 0 {
		changeType = internal.REGRESSION
	} else if flatChange < 0 {
		changeType = internal.IMPROVEMENT
	}

	return &FunctionChangeResult{
		FunctionName:      current.FnName,
		ChangeType:        changeType,
		FlatChangePercent: flatChange,
		CumChangePercent:  cumChange,
		FlatAbsolute: AbsoluteChange{
			Before: baseline.Flat,
			After:  current.Flat,
			Delta:  current.Flat - baseline.Flat,
		},
		CumAbsolute: AbsoluteChange{
			Before: baseline.Cum,
			After:  current.Cum,
			Delta:  current.Cum - baseline.Cum,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cr *FunctionChangeResult) writeHeader(report *strings.Builder) {
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString("               PERFORMANCE CHANGE REPORT\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
}

func (cr *FunctionChangeResult) writeFunctionInfo(report *strings.Builder) {
	fmt.Fprintf(report, "Function: %s\n", cr.FunctionName)
	fmt.Fprintf(report, "Analysis Time: %s\n", cr.Timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(report, "Change Type: %s\n\n", cr.ChangeType)
}

func (cr *FunctionChangeResult) writeStatusAssessment(report *strings.Builder) {
	statusIcon := map[string]string{
		internal.IMPROVEMENT: "✅",
		internal.REGRESSION:  "⚠️",
	}[cr.ChangeType]

	if statusIcon == "" {
		statusIcon = "🔄"
	}

	assessment := map[string]string{
		internal.IMPROVEMENT: "Performance improvement detected",
		internal.REGRESSION:  "Performance regression detected",
	}[cr.ChangeType]

	if assessment == "" {
		assessment = "No significant change detected"
	}

	fmt.Fprintf(report, "%s %s\n\n", statusIcon, assessment)
}

func (cr *FunctionChangeResult) writeFlatAnalysis(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    FLAT TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	sign := signPrefix(cr.FlatChangePercent)

	fmt.Fprintf(report, "Before:       %.6fs\n", cr.FlatAbsolute.Before)
	fmt.Fprintf(report, "After:        %.6fs\n", cr.FlatAbsolute.After)
	fmt.Fprintf(report, "Delta:        %s%.6fs\n", sign, cr.FlatAbsolute.Delta)
	fmt.Fprintf(report, "Change:       %s%.2f%%\n", sign, cr.FlatChangePercent)

	switch {
	case cr.FlatChangePercent > 0:
		fmt.Fprintf(report, "Impact:       Function is %.2f%% SLOWER\n\n", cr.FlatChangePercent)
	case cr.FlatChangePercent < 0:
		fmt.Fprintf(report, "Impact:       Function is %.2f%% FASTER\n\n", math.Abs(cr.FlatChangePercent))
	default:
		report.WriteString("Impact:       No change in execution time\n\n")
	}
}

func (cr *FunctionChangeResult) writeCumulativeAnalysis(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                 CUMULATIVE TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	sign := signPrefix(cr.CumChangePercent)

	fmt.Fprintf(report, "Before:       %.3fs\n", cr.CumAbsolute.Before)
	fmt.Fprintf(report, "After:        %.3fs\n", cr.CumAbsolute.After)
	fmt.Fprintf(report, "Delta:        %s%.3fs\n", sign, cr.CumAbsolute.Delta)
	fmt.Fprintf(report, "Change:       %s%.2f%%\n\n", sign, cr.CumChangePercent)
}

func (cr *FunctionChangeResult) writeImpactAssessment(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    IMPACT ASSESSMENT\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	fmt.Fprintf(report, "Severity:     %s\n", cr.calculateSeverity())
	report.WriteString("Recommendation: ")
	report.WriteString(cr.recommendation())
	report.WriteString("\n")
}

func getBinFilesLocations(selections *Selections) (string, string) {
	fileName := fmt.Sprintf("%s_%s.out", selections.BenchmarkName, selections.ProfileType)
	binFilePath1BaseLine := filepath.Join(internal.MainDirOutput, selections.Baseline, internal.ProfileBinDir, selections.BenchmarkName, fileName)
	binFilePath2Current := filepath.Join(internal.MainDirOutput, selections.Current, internal.ProfileBinDir, selections.BenchmarkName, fileName)

	return binFilePath1BaseLine, binFilePath2Current
}

func chooseFileLocations(selections *Selections) (string, string) {
	var textFilePathBaseLine, textFilePathCurrent string

	if selections.IsManual {
		textFilePathBaseLine = selections.Baseline
		textFilePathCurrent = selections.Current
	} else {
		textFilePathBaseLine, textFilePathCurrent = getBinFilesLocations(selections)
	}

	return textFilePathBaseLine, textFilePathCurrent
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
		if worst != nil && worst.FlatChangePercent <= effectiveThreshold {
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

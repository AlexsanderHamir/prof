package tracker

import (
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/AlexsanderHamir/prof/internal"
)

func applyCIConfiguration(report *ProfileChangeReport, selections *Selections) error {
	cfg, err := internal.LoadFromFile(internal.ConfigFilename)
	if err != nil {
		slog.Info("No CI/CD configuration found, using command-line settings only")
		return applyCommandLineThresholds(report, selections)
	}

	report.ApplyCIConfiguration(cfg.CIConfig, selections.BenchmarkName)

	cliFlagsProvided := selections.UseThreshold || selections.RegressionThreshold > 0.0
	if cliFlagsProvided {
		return applyCommandLineThresholds(report, selections)
	}

	slog.Info("No CLI regression flags provided, using CI/CD configuration settings")
	return applyCICDThresholdsOnly(report, selections, cfg.CIConfig)
}

func applyCommandLineThresholds(report *ProfileChangeReport, selections *Selections) error {
	if selections.UseThreshold && selections.RegressionThreshold > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent >= selections.RegressionThreshold {
			return fmt.Errorf("performance regression %.2f%% in %s exceeds threshold %.2f%%",
				worst.FlatChangePercent, worst.FunctionName, selections.RegressionThreshold)
		}
	}
	return nil
}

func applyCICDThresholdsOnly(report *ProfileChangeReport, selections *Selections, cicdConfig *internal.CIConfig) error {
	benchmarkName := selections.BenchmarkName
	if benchmarkName == "" {
		benchmarkName = "unknown"
	}

	effectiveThreshold := getEffectiveRegressionThreshold(cicdConfig, benchmarkName, 0.0)
	minChangeThreshold := getMinChangeThreshold(cicdConfig, benchmarkName)
	failOnImprovement := shouldFailOnImprovement(cicdConfig, benchmarkName)

	if effectiveThreshold > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent <= effectiveThreshold {
			if !shouldIgnoreFunction(cicdConfig, worst.FunctionName, benchmarkName) {
				return fmt.Errorf("performance regression %.2f%% in %s exceeds CI/CD threshold %.2f%%",
					worst.FlatChangePercent, worst.FunctionName, effectiveThreshold)
			}
		}
	}

	if failOnImprovement {
		best := report.BestImprovement()
		if best != nil && math.Abs(best.FlatChangePercent) >= minChangeThreshold {
			if !shouldIgnoreFunction(cicdConfig, best.FunctionName, benchmarkName) {
				return fmt.Errorf("unexpected performance improvement %.2f%% in %s (configured to fail on improvements)",
					math.Abs(best.FlatChangePercent), best.FunctionName)
			}
		}
	}

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

func getEffectiveRegressionThreshold(cicdConfig *internal.CIConfig, benchmarkName string, commandLineThreshold float64) float64 {
	if cicdConfig == nil {
		return commandLineThreshold
	}
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists && benchmarkConfig.MaxRegressionThreshold > 0 {
		if commandLineThreshold > 0 {
			if commandLineThreshold < benchmarkConfig.MaxRegressionThreshold {
				return commandLineThreshold
			}
			return benchmarkConfig.MaxRegressionThreshold
		}
		return benchmarkConfig.MaxRegressionThreshold
	}
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
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists && benchmarkConfig.MinChangeThreshold > 0 {
		return benchmarkConfig.MinChangeThreshold
	}
	if cicdConfig.Global != nil && cicdConfig.Global.MinChangeThreshold > 0 {
		return cicdConfig.Global.MinChangeThreshold
	}
	return 0.0
}

func shouldFailOnImprovement(cicdConfig *internal.CIConfig, benchmarkName string) bool {
	if cicdConfig == nil {
		return false
	}
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists {
		return benchmarkConfig.FailOnImprovement
	}
	if cicdConfig.Global != nil {
		return cicdConfig.Global.FailOnImprovement
	}
	return false
}

func shouldIgnoreFunction(cicdConfig *internal.CIConfig, functionName string, benchmarkName string) bool {
	if cicdConfig == nil {
		return false
	}
	if benchmarkConfig, exists := cicdConfig.Benchmarks[benchmarkName]; exists {
		if shouldIgnoreFunctionByConfig(&benchmarkConfig, functionName) {
			return true
		}
	}
	if cicdConfig.Global != nil {
		return shouldIgnoreFunctionByConfig(cicdConfig.Global, functionName)
	}
	return false
}

func shouldIgnoreFunctionByConfig(config *internal.CITrackingConfig, functionName string) bool {
	if config == nil {
		return false
	}
	for _, ignoredFunc := range config.IgnoreFunctions {
		if functionName == ignoredFunc {
			return true
		}
	}
	for _, ignoredPrefix := range config.IgnorePrefixes {
		if strings.HasPrefix(functionName, ignoredPrefix) {
			return true
		}
	}
	return false
}

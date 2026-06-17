package tracker

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/AlexsanderHamir/prof/internal/config"
)

func applyCIConfiguration(report *ProfileChangeReport, selections *Selections) error {
	cfg, err := config.Load()
	if err != nil {
		slog.Info("No track configuration found, using command-line settings only")
		return applyCommandLineThresholds(report, selections)
	}

	benchmarkName := selections.BenchmarkName
	if benchmarkName == "" {
		benchmarkName = "unknown"
	}

	policy := config.ResolveTrackPolicy(cfg, benchmarkName)
	report.ApplyTrackPolicy(policy)

	cliFlagsProvided := selections.UseThreshold || selections.RegressionThreshold > 0.0
	if cliFlagsProvided {
		return applyCommandLineThresholds(report, selections)
	}

	slog.Info("No CLI regression flags provided, using track configuration settings")
	return applyTrackThresholdsOnly(report, policy)
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

func applyTrackThresholdsOnly(report *ProfileChangeReport, policy config.TrackPolicy) error {
	if policy.MaxRegressionPercent > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent >= policy.MaxRegressionPercent {
			if !config.ShouldIgnoreFunction(policy, worst.FunctionName) {
				return fmt.Errorf("performance regression %.2f%% in %s exceeds track threshold %.2f%%",
					worst.FlatChangePercent, worst.FunctionName, policy.MaxRegressionPercent)
			}
		}
	}

	if policy.FailOnImprovement {
		best := report.BestImprovement()
		if best != nil && math.Abs(best.FlatChangePercent) >= policy.MinChangePercent {
			if !config.ShouldIgnoreFunction(policy, best.FunctionName) {
				return fmt.Errorf("unexpected performance improvement %.2f%% in %s (configured to fail on improvements)",
					math.Abs(best.FlatChangePercent), best.FunctionName)
			}
		}
	}

	if policy.MinChangePercent > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent < policy.MinChangePercent {
			slog.Info("Performance regression below minimum threshold, not failing",
				"function", worst.FunctionName,
				"change", worst.FlatChangePercent,
				"threshold", policy.MinChangePercent)
		}
	}

	return nil
}

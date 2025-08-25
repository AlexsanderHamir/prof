package cli

import (
	"github.com/AlexsanderHamir/prof/engine/tracker"
)

// setGlobalTrackingVariables sets the global CLI variables for tracking
func setGlobalTrackingVariables(selections *tracker.Selections) {
	Baseline = selections.Baseline
	Current = selections.Current
	benchmarkName = selections.BenchmarkName
	profileType = selections.ProfileType
	outputFormat = selections.OutputFormat
	failOnRegression = selections.UseThreshold
	regressionThreshold = selections.RegressionThreshold
}

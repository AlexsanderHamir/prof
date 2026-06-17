package cli

import (
	"github.com/AlexsanderHamir/prof/internal/app"
)

// setGlobalTrackingVariables copies track selections into tools flags for legacy wiring.
func setGlobalTrackingVariables(selections *app.TrackOptions) {
	f := &toolsGlobal
	f.baseline = selections.Baseline
	f.current = selections.Current
	f.benchmarkName = selections.BenchmarkName
	f.profileType = selections.ProfileType
}

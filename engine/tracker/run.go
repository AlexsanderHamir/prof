package tracker

import (
	"fmt"
	"log/slog"
)

var validFormats = map[string]bool{
	"summary":       true,
	"detailed":      true,
	"summary-html":  true,
	"detailed-html": true,
	"summary-json":  true,
	"detailed-json": true,
}

// RunTrackAuto compares runs collected via prof auto (tag layout).
func RunTrackAuto(selections *Selections) error {
	return runTrack(selections)
}

// RunTrackManual compares profile text files at paths given in Baseline / Current.
func RunTrackManual(selections *Selections) error {
	return runTrack(selections)
}

func runTrack(selections *Selections) error {
	if !validFormats[selections.OutputFormat] {
		return fmt.Errorf("invalid output format '%s'. Valid formats: summary, detailed", selections.OutputFormat)
	}

	report, err := CheckPerformanceDifferences(selections)
	if err != nil {
		return fmt.Errorf("failed to track performance differences: %w", err)
	}

	if len(report.FunctionChanges) == 0 {
		slog.Info("No function changes detected between the two runs")
		return nil
	}

	report.ChooseOutputFormat(selections.OutputFormat)
	return applyCIConfiguration(report, selections)
}

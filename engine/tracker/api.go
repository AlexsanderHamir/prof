package tracker

import (
	"fmt"
	"log/slog"
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

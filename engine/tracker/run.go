package tracker

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

var validFormats = map[string]bool{
	"summary":       true,
	"detailed":      true,
	"summary-html":  true,
	"detailed-html": true,
	"summary-json":  true,
	"detailed-json": true,
}

func validFormatNames() []string {
	names := make([]string, 0, len(validFormats))
	for k := range validFormats {
		names = append(names, k)
	}
	slices.Sort(names)
	return names
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
		return fmt.Errorf("invalid output format %q (valid: %s)",
			selections.OutputFormat,
			strings.Join(validFormatNames(), ", "))
	}

	report, err := CheckPerformanceDifferences(selections)
	if err != nil {
		return fmt.Errorf("failed to track performance differences: %w", err)
	}

	if len(report.FunctionChanges) == 0 {
		slog.Info("No function changes detected between the two runs")
		return nil
	}

	if err := report.ChooseOutputFormat(selections.OutputFormat); err != nil {
		return fmt.Errorf("write report (%s): %w", selections.OutputFormat, err)
	}
	return applyCIConfiguration(report, selections)
}

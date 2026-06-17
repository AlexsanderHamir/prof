package tracker

import (
	"fmt"
	"log/slog"
	"strings"
)

// RunTrackAuto compares runs collected via prof auto (tag layout).
func RunTrackAuto(opts Options) error {
	opts.IsManual = false
	return runTrack(&opts)
}

// RunTrackManual compares binary pprof profiles at paths in Baseline / Current.
func RunTrackManual(opts Options) error {
	opts.IsManual = true
	return runTrack(&opts)
}

func runTrack(selections *Options) error {
	if !ValidOutputFormat(selections.OutputFormat) {
		return fmt.Errorf("invalid output format %q (valid: %s)",
			selections.OutputFormat,
			strings.Join(ValidOutputFormats, ", "))
	}

	report, err := CheckPerformanceDifferences(selections)
	if err != nil {
		return fmt.Errorf("failed to track performance differences: %w", err)
	}

	if len(report.FunctionChanges) == 0 {
		slog.Info("No function changes detected between the two runs")
		return nil
	}

	if formatErr := report.ChooseOutputFormat(selections.OutputFormat); formatErr != nil {
		return fmt.Errorf("write report (%s): %w", selections.OutputFormat, formatErr)
	}
	return applyCIConfiguration(report, selections)
}

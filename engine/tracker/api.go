package tracker

import (
	"fmt"
	"log/slog"

	"github.com/AlexsanderHamir/prof/parser"
)

// trackAutoSelections holds all the user selections for tracking
type Selections struct {
	BaselineTag         string
	CurrentTag          string
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

	if selections.UseThreshold && selections.RegressionThreshold > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent >= selections.RegressionThreshold {
			return fmt.Errorf("performance regression %.2f%% in %s exceeds threshold %.2f%%", worst.FlatChangePercent, worst.FunctionName, selections.RegressionThreshold)
		}
	}

	return nil
}

// RunTrackManual receives the location of the .out / .prof files,
// and does what RunTrackAuto does.
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

	if selections.UseThreshold && selections.RegressionThreshold > 0.0 {
		worst := report.WorstRegression()
		if worst != nil && worst.FlatChangePercent >= selections.RegressionThreshold {
			return fmt.Errorf("performance regression %.2f%% in %s exceeds threshold %.2f%%", worst.FlatChangePercent, worst.FunctionName, selections.RegressionThreshold)
		}
	}

	return nil
}

// CheckPerformanceDifferences creates the profile report by comparing data from  prof's auto run.
func CheckPerformanceDifferences(selections *Selections) (*ProfileChangeReport, error) {
	textFilePathBaseLine, textFilePathCurrent := chooseFileLocations(selections)

	lineObjsBaseline, err := parser.TurnLinesIntoObjects(textFilePathBaseLine)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", textFilePathBaseLine, err)
	}

	lineObjsCurrent, err := parser.TurnLinesIntoObjects(textFilePathCurrent)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", textFilePathCurrent, err)
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

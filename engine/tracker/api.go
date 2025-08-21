package tracker

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

// trackAutoSelections holds all the user selections for tracking
type AutoSelections struct {
	BaselineTag         string
	CurrentTag          string
	BenchmarkName       string
	ProfileType         string
	OutputFormat        string
	UseThreshold        bool
	RegressionThreshold float64
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
func RunTrackAuto(selections *AutoSelections) error {
	if !validFormats[selections.OutputFormat] {
		return fmt.Errorf("invalid output format '%s'. Valid formats: summary, detailed", selections.OutputFormat)
	}

	report, err := CheckPerformanceDifferences(selections.BaselineTag, selections.CurrentTag, selections.BenchmarkName, selections.ProfileType)
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

func RunTrackManual(selections *AutoSelections) error {
	if !validFormats[selections.OutputFormat] {
		return fmt.Errorf("invalid output format '%s'. Valid formats: summary, detailed", selections.OutputFormat)
	}

	report, err := CheckPerformanceDifferencesManual(selections.BaselineTag, selections.CurrentTag)
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
func CheckPerformanceDifferences(baselineTag, currentTag, benchName, profileType string) (*ProfileChangeReport, error) {
	fileName := fmt.Sprintf("%s_%s.txt", benchName, profileType)
	textFilePath1BaseLine := filepath.Join(internal.MainDirOutput, baselineTag, internal.ProfileTextDir, benchName, fileName)
	textFilePath2Current := filepath.Join(internal.MainDirOutput, currentTag, internal.ProfileTextDir, benchName, fileName)

	lineObjsBaseline, err := parser.TurnLinesIntoObjects(textFilePath1BaseLine)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", textFilePath1BaseLine, err)
	}

	lineObjsCurrent, err := parser.TurnLinesIntoObjects(textFilePath2Current)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", textFilePath2Current, err)
	}

	matchingMap := createHashFromLineObjects(lineObjsBaseline)

	pgp := &ProfileChangeReport{}
	for _, currentObj := range lineObjsCurrent {
		baseLineObj, matchNotFound := matchingMap[currentObj.FnName]
		if !matchNotFound {
			continue
		}

		var changeResult *FunctionChangeResult
		changeResult, err = DetectChange(baseLineObj, currentObj)
		if err != nil {
			return nil, fmt.Errorf("DetectChange failed: %w", err)
		}

		pgp.FunctionChanges = append(pgp.FunctionChanges, changeResult)
	}

	return pgp, nil
}

// CheckPerformanceDifferences creates the profile report by comparing data from  prof's auto run.
func CheckPerformanceDifferencesManual(baselineProfile, currentProfile string) (*ProfileChangeReport, error) {
	lineObjsBaseline, err := parser.TurnLinesIntoObjects(baselineProfile)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", baselineProfile, err)
	}

	lineObjsCurrent, err := parser.TurnLinesIntoObjects(currentProfile)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", currentProfile, err)
	}

	matchingMap := createHashFromLineObjects(lineObjsBaseline)

	pgp := &ProfileChangeReport{}
	for _, currentObj := range lineObjsCurrent {
		baseLineObj, matchNotFound := matchingMap[currentObj.FnName]
		if !matchNotFound {
			continue
		}

		var changeResult *FunctionChangeResult
		changeResult, err = DetectChange(baseLineObj, currentObj)
		if err != nil {
			return nil, fmt.Errorf("DetectChange failed: %w", err)
		}

		pgp.FunctionChanges = append(pgp.FunctionChanges, changeResult)
	}

	return pgp, nil
}

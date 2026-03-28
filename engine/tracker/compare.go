package tracker

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func detectChangeBetweenTwoObjects(baseline, current *parser.LineObj) (*FunctionChangeResult, error) {
	if current == nil {
		return nil, errors.New("current obj is nil")
	}
	if baseline == nil {
		return nil, errors.New("baseLine obj is nil")
	}
	const percentMultiplier = 100
	var flatChange float64
	if baseline.Flat != 0 {
		flatChange = ((current.Flat - baseline.Flat) / baseline.Flat) * percentMultiplier
	}
	var cumChange float64
	if baseline.Cum != 0 {
		cumChange = ((current.Cum - baseline.Cum) / baseline.Cum) * percentMultiplier
	}
	changeType := internal.STABLE
	if flatChange > 0 {
		changeType = internal.REGRESSION
	} else if flatChange < 0 {
		changeType = internal.IMPROVEMENT
	}
	return &FunctionChangeResult{
		FunctionName:      current.FnName,
		ChangeType:        changeType,
		FlatChangePercent: flatChange,
		CumChangePercent:  cumChange,
		FlatAbsolute: AbsoluteChange{
			Before: baseline.Flat,
			After:  current.Flat,
			Delta:  current.Flat - baseline.Flat,
		},
		CumAbsolute: AbsoluteChange{
			Before: baseline.Cum,
			After:  current.Cum,
			Delta:  current.Cum - baseline.Cum,
		},
		Timestamp: time.Now(),
	}, nil
}

func lineObjByShortName(lineobjects []*parser.LineObj) map[string]*parser.LineObj {
	m := make(map[string]*parser.LineObj)
	for _, o := range lineobjects {
		m[o.FnName] = o
	}
	return m
}

func getBinFilesLocations(selections *Selections) (string, string) {
	fileName := fmt.Sprintf("%s_%s.out", selections.BenchmarkName, selections.ProfileType)
	base := filepath.Join(internal.MainDirOutput, selections.Baseline, internal.ProfileBinDir, selections.BenchmarkName, fileName)
	cur := filepath.Join(internal.MainDirOutput, selections.Current, internal.ProfileBinDir, selections.BenchmarkName, fileName)
	return base, cur
}

func chooseFileLocations(selections *Selections) (baselinePath, currentPath string) {
	if selections.IsManual {
		return selections.Baseline, selections.Current
	}
	return getBinFilesLocations(selections)
}

// CheckPerformanceDifferences loads both profiles and pairs functions by short name.
func CheckPerformanceDifferences(selections *Selections) (*ProfileChangeReport, error) {
	binFilePathBaseLine, binFilePathCurrent := chooseFileLocations(selections)

	lineObjsBaseline, err := parser.TurnLinesIntoObjectsV2(binFilePathBaseLine)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", binFilePathBaseLine, err)
	}

	lineObjsCurrent, err := parser.TurnLinesIntoObjectsV2(binFilePathCurrent)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", binFilePathCurrent, err)
	}

	byName := lineObjByShortName(lineObjsBaseline)
	report := &ProfileChangeReport{}
	for _, currentObj := range lineObjsCurrent {
		baseLineObj, ok := byName[currentObj.FnName]
		if !ok {
			continue
		}
		changeResult, derr := detectChangeBetweenTwoObjects(baseLineObj, currentObj)
		if derr != nil {
			return nil, fmt.Errorf("detectChangeBetweenTwoObjects failed: %w", derr)
		}
		report.FunctionChanges = append(report.FunctionChanges, changeResult)
	}
	return report, nil
}

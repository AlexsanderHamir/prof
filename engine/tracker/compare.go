package tracker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlexsanderHamir/prof/internal/workspace"
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
	changeType := ChangeStable
	if flatChange > 0 {
		changeType = ChangeRegression
	} else if flatChange < 0 {
		changeType = ChangeImprovement
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

func getBinFilesLocations(selections *Options) (string, string) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	baseRoot := filepath.Join(cwd, workspace.MainDirOutput, selections.Baseline)
	curRoot := filepath.Join(cwd, workspace.MainDirOutput, selections.Current)
	baseLayout := workspace.TagLayout{Tag: selections.Baseline, Root: baseRoot}
	curLayout := workspace.TagLayout{Tag: selections.Current, Root: curRoot}
	return baseLayout.Bin(selections.BenchmarkName, selections.ProfileType),
		curLayout.Bin(selections.BenchmarkName, selections.ProfileType)
}

func chooseFileLocations(selections *Options) (baselinePath, currentPath string) {
	if selections.IsManual {
		return selections.Baseline, selections.Current
	}
	return getBinFilesLocations(selections)
}

// CheckPerformanceDifferences loads both profiles and pairs functions by short name.
func CheckPerformanceDifferences(selections *Options) (*ProfileChangeReport, error) {
	binFilePathBaseLine, binFilePathCurrent := chooseFileLocations(selections)

	lineObjsBaseline, err := loadProfileObjects(binFilePathBaseLine, selections, "baseline")
	if err != nil {
		return nil, err
	}

	lineObjsCurrent, err := loadProfileObjects(binFilePathCurrent, selections, "current")
	if err != nil {
		return nil, err
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

func loadProfileObjects(path string, selections *Options, role string) ([]*parser.LineObj, error) {
	objs, err := parser.TurnLinesIntoObjectsV2(path)
	if err == nil {
		return objs, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		tag := selections.Baseline
		if role == "current" {
			tag = selections.Current
		}
		return nil, fmt.Errorf("%s tag %q has no %s/%s profile — run collect for that benchmark or pick another",
			role, tag, selections.BenchmarkName, selections.ProfileType)
	}
	return nil, fmt.Errorf("couldn't load profile at %s: %w", path, err)
}

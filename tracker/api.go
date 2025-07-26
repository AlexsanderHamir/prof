package tracker

import (
	"fmt"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
)

// CheckPerformanceDifferences creates the profile report by comparing data from  prof's auto run.
func CheckPerformanceDifferences(baselineTag, currentTag, benchName, profileType string) (*ProfileChangeReport, error) {
	fileName := fmt.Sprintf("%s_%s.txt", benchName, profileType)
	textFilePath1BaseLine := filepath.Join(shared.MainDirOutput, baselineTag, shared.ProfileTextDir, benchName, fileName)
	textFilePath2Current := filepath.Join(shared.MainDirOutput, currentTag, shared.ProfileTextDir, benchName, fileName)

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

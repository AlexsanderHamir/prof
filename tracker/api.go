package tracker

import (
	"fmt"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
)

func CheckPerformanceDifferences(tagPath1, tagPath2, benchName, profileType string) (*ProfileChangeReport, error) {
	fileName := fmt.Sprintf("%s_%s.txt", benchName, profileType)
	textFilePath1 := filepath.Join(tagPath1, shared.ProfileTextDir, benchName, fileName)
	textFilePath2 := filepath.Join(tagPath2, shared.ProfileTextDir, benchName, fileName)

	lineObjs1, err := parser.TurnLinesIntoObjects(textFilePath1, profileType)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", textFilePath1, err)
	}

	lineObjs2, err := parser.TurnLinesIntoObjects(textFilePath2, profileType)
	if err != nil {
		return nil, fmt.Errorf("couldn't get objs for path: %s, error: %w", textFilePath1, err)
	}

	matchingMap := createHashFromLineObjects(lineObjs1)

	pgp := &ProfileChangeReport{}
	for _, current := range lineObjs2 {
		baseLine, matchNotFound := matchingMap[current.FnName]
		if !matchNotFound {
			continue
		}

		changeResult, err := DetectChange(baseLine, current)
		if err != nil {
			return nil, fmt.Errorf("DetectChange failed: %w", err)
		}

		pgp.FunctionChanges = append(pgp.FunctionChanges, changeResult)
	}

	return pgp, nil
}

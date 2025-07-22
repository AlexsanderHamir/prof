package regressor

import (
	"fmt"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
)

func CheckPerformanceDifferences(tagPath1, tagPath2, benchName, profileType string) error {
	fileName := fmt.Sprintf("%s_%s.txt", benchName, profileType)
	textFilePath1 := filepath.Join(tagPath1, shared.ProfileTextDir, benchName, fileName)
	textFilePath2 := filepath.Join(tagPath2, shared.ProfileTextDir, benchName, fileName)

	lineObjs1, err := parser.TurnLinesIntoObjects(textFilePath1, profileType)
	if err != nil {
		return fmt.Errorf("couldn't get objs for path: %s", textFilePath1)
	}

	lineObjs2, err := parser.TurnLinesIntoObjects(textFilePath2, profileType)
	if err != nil {
		return fmt.Errorf("couldn't get objs for path: %s", textFilePath2)
	}

	matchingMap := createHashFromLineObjects(lineObjs1)

	for i, current := range lineObjs2 {
		if i == 0 {
			continue
		}
		baseLine := matchingMap[current.FnName]
		changeResult := DetectChange(baseLine, current)
		res := changeResult.Report()
		fmt.Println(res)
		panic("Stop")
	}

	return nil
}

func createHashFromLineObjects(lineobjects []*parser.LineObj) map[string]*parser.LineObj {
	matchingMap := make(map[string]*parser.LineObj)
	for _, lineObj := range lineobjects {
		matchingMap[lineObj.FnName] = lineObj
	}

	return matchingMap
}

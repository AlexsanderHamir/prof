package regressor

import (
	"fmt"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
)

// Basic building block: comparing two tags
// Create black box test for this

// 1. Receive two tags
// 2. Get all data
// 3. Get full function name
// 4. Compare

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

	fmt.Println("one: ", lineObjs1)
	fmt.Println("two: ", lineObjs2)

	return nil
}

package parser

import (
	"github.com/AlexsanderHamir/prof/internal"
)

// TurnLinesIntoObjectsV2 turn profile data from a .pprof file into line objects.
func TurnLinesIntoObjectsV2(profilePath string) ([]*LineObj, error) {
	profileData, err := extractProfileData(profilePath)
	if err != nil {
		return nil, err
	}

	var lineObjs []*LineObj
	for _, entry := range profileData.SortedEntries {
		fn := entry.Name
		lineObj := &LineObj{
			FnName:         fn,
			Flat:           float64(entry.Flat),
			FlatPercentage: profileData.FlatPercentages[fn],
			SumPercentage:  profileData.SumPercentages[fn],
			Cum:            float64(profileData.Cum[fn]),
			CumPercentage:  profileData.CumPercentages[fn],
		}
		lineObjs = append(lineObjs, lineObj)
	}

	return lineObjs, nil
}

// GetAllFunctionNamesV2 extracts all function names from a profile (.pprof) file.
func GetAllFunctionNamesV2(profilePath string, filter internal.FunctionFilter) (names []string, err error) {
	profileData, err := extractProfileData(profilePath)
	if err != nil {
		return nil, err
	}

	ignoreSet := getFilterSets(filter.IgnoreFunctions)
	for _, entry := range profileData.SortedEntries {
		fn := entry.Name

		// Extract the function name from the full function path
		funcName := extractSimpleFunctionName(fn)
		if funcName == "" {
			continue
		}

		// Check if function should be ignored
		if _, ignored := ignoreSet[funcName]; ignored {
			continue
		}

		// Check if function matches include prefixes
		if len(filter.IncludePrefixes) > 0 && !matchPrefix(fn, filter.IncludePrefixes) {
			continue
		}

		names = append(names, funcName)
	}

	return names, nil
}

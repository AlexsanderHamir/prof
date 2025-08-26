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

// GetAllFunctionNamesV2 extracts all function names from a pprof file, the function name is the name after the last dot.
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

// OrganizeProfileByPackageV2 organizes profile data by package/module and returns a formatted string
// that groups functions by their package/module with subtotals and percentages.
func OrganizeProfileByPackageV2(profilePath string, filter internal.FunctionFilter) (string, error) {
	profileData, err := extractProfileData(profilePath)
	if err != nil {
		return "", err
	}

	// Group functions by package/module
	packageGroups := make(map[string]*PackageGroup)
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

		// Extract package name
		packageName := extractPackageName(fn)
		if packageName == "" {
			packageName = "unknown"
		}

		// Initialize package group if it doesn't exist
		if packageGroups[packageName] == nil {
			packageGroups[packageName] = &PackageGroup{
				Name:      packageName,
				Functions: make([]*FunctionInfo, 0),
				TotalFlat: 0,
				TotalCum:  0,
			}
		}

		// Add function to package group
		funcInfo := &FunctionInfo{
			Name:           funcName,
			FullName:       fn,
			Flat:           float64(entry.Flat),
			FlatPercentage: profileData.FlatPercentages[fn],
			Cum:            float64(profileData.Cum[fn]),
			CumPercentage:  profileData.CumPercentages[fn],
			SumPercentage:  profileData.SumPercentages[fn],
		}

		packageGroups[packageName].Functions = append(packageGroups[packageName].Functions, funcInfo)
		packageGroups[packageName].TotalFlat += funcInfo.Flat
		packageGroups[packageName].TotalCum += funcInfo.Cum
	}

	// Calculate package percentages and sort
	totalFlat := float64(profileData.Total)
	for _, pkg := range packageGroups {
		pkg.FlatPercentage = pkg.TotalFlat / totalFlat * 100
		pkg.CumPercentage = pkg.TotalCum / totalFlat * 100
	}

	// Sort packages by flat percentage (descending)
	sortedPackages := sortPackagesByFlatPercentage(packageGroups)

	// Generate formatted output
	return formatPackageReport(sortedPackages), nil
}

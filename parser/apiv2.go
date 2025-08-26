package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlexsanderHamir/prof/internal"
)

type LineObj struct {
	FnName         string
	Flat           float64
	FlatPercentage float64
	SumPercentage  float64
	Cum            float64
	CumPercentage  float64
}

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

// PackageGroup represents a group of functions from the same package
type PackageGroup struct {
	Name           string
	Functions      []*FunctionInfo
	TotalFlat      float64
	TotalCum       float64
	FlatPercentage float64
	CumPercentage  float64
}

// FunctionInfo represents a function with its performance metrics
type FunctionInfo struct {
	Name           string
	FullName       string
	Flat           float64
	FlatPercentage float64
	Cum            float64
	CumPercentage  float64
}

// extractPackageName extracts the package name from a full function path
func extractPackageName(fullPath string) string {
	// Handle cases like "github.com/user/pkg.(*Type).Method" => "github.com/user/pkg"
	// or "sync/atomic.CompareAndSwapPointer" => "sync/atomic"

	// Split by dots
	parts := strings.Split(fullPath, ".")
	if len(parts) < 2 {
		return ""
	}

	// Check if it's a standard library package (like "sync/atomic")
	if !strings.Contains(parts[0], "/") && len(parts) >= 2 {
		// Standard library package
		if len(parts) >= 3 && strings.Contains(parts[1], "/") {
			return parts[0] + "." + parts[1]
		}
		return parts[0]
	}

	// Check if it's a GitHub-style package
	if strings.Contains(parts[0], "github.com") || strings.Contains(parts[0], "golang.org") {
		// For GitHub packages, take up to the third part (github.com/user/pkg)
		if len(parts) >= 3 {
			return strings.Join(parts[:3], ".")
		}
		return strings.Join(parts[:2], ".")
	}

	// For other cases, take the first part
	return parts[0]
}

// sortPackagesByFlatPercentage sorts packages by their flat percentage in descending order
func sortPackagesByFlatPercentage(packageGroups map[string]*PackageGroup) []*PackageGroup {
	var packages []*PackageGroup
	for _, pkg := range packageGroups {
		packages = append(packages, pkg)
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].FlatPercentage > packages[j].FlatPercentage
	})

	return packages
}

// formatPackageReport formats the package groups into a readable report
func formatPackageReport(packages []*PackageGroup) string {
	var result strings.Builder

	for i, pkg := range packages {
		if i > 0 {
			result.WriteString("\n\n")
		}

		// Package header
		result.WriteString(fmt.Sprintf("#### **%s**\n", pkg.Name))

		// Sort functions by flat percentage (descending)
		sort.Slice(pkg.Functions, func(i, j int) bool {
			return pkg.Functions[i].FlatPercentage > pkg.Functions[j].FlatPercentage
		})

		// List functions
		for _, fn := range pkg.Functions {
			if fn.Flat > 0 {
				// Show only function name and percentage
				result.WriteString(fmt.Sprintf("- `%s` → %.2f%%\n",
					fn.Name, fn.FlatPercentage))
			} else if fn.Cum > 0 {
				// Function with only cumulative time
				result.WriteString(fmt.Sprintf("- `%s` → 0%% (cum %.2f%%)\n",
					fn.Name, fn.CumPercentage))
			}
		}

		// Package subtotal
		result.WriteString(fmt.Sprintf("\n**Subtotal (%s)**: ≈%.1f%%",
			extractShortPackageName(pkg.Name), pkg.FlatPercentage))
	}

	return result.String()
}

// extractShortPackageName extracts a shorter version of the package name for display
func extractShortPackageName(fullPackageName string) string {
	parts := strings.Split(fullPackageName, ".")
	if len(parts) == 0 {
		return fullPackageName
	}

	// For GitHub packages, show just the last part
	if strings.Contains(fullPackageName, "github.com") {
		return parts[len(parts)-1]
	}

	// For standard library, show the full name
	return fullPackageName
}

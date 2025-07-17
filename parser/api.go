package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/shared"
)

var (
	funcNameRegexp = regexp.MustCompile(`\.([^.(]+)(?:\([^)]*\))?$`)
	floatRegexp    = regexp.MustCompile(`\d+(?:\.\d+)?`)
	header         = "flat  flat%   sum%        cum   cum%"
)

// ProfileFilter collects filters for extracting function names from a profile.
type ProfileFilter struct {
	// Include only lines starting with specified prefix
	FunctionPrefixes []string

	// Ignore all functions after the last dot even if includes the above prefix
	IgnoreFunctions []string
}

// GetAllFunctionNames extracts all function names from a profile text file, applying the given filter.
func GetAllFunctionNames(filePath string, filter ProfileFilter) ([]string, error) {
	scanner, file, err := shared.GetScanner(filePath)
	if err != nil {
		return nil, fmt.Errorf("GetAllFunctionNames Failed: %w", err)
	}
	defer file.Close()

	ignoreSet := getFilterSets(filter.IgnoreFunctions)

	var functions []string
	var foundHeader bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, header) {
			foundHeader = true
			continue
		}

		// Skip lines until we find the header, then process profile data
		if !foundHeader {
			continue
		}

		if funcName := extractFunctionName(line, filter.FunctionPrefixes, ignoreSet); funcName != "" {
			functions = append(functions, funcName)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading profile file: %w", err)
	}

	if !foundHeader {
		return nil, fmt.Errorf("profile file header not found")
	}

	return functions, nil
}

// ShouldKeepLine determines if a line from a profile should be kept based on profile values and ignore filters.
func ShouldKeepLine(line string, profileFilters map[int]float64, ignoreFunctionSet, ignorePrefixSet map[string]struct{}) bool {
	if line == "" {
		return false
	}

	parts := strings.Fields(line)
	if len(parts) < 6 {
		return false
	}

	if !filterByNumber(profileFilters, parts) {
		return false
	}

	if !filterByIgnoreFunctions(ignoreFunctionSet, parts) {
		return false
	}

	return filterByIgnorePrefixes(ignorePrefixSet, parts)
}

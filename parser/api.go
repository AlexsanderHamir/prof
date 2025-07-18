package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/shared"
)

const (
	funcNameRegexp = `\.([^.(]+)(?:\([^)]*\))?$`
	floatRegexp    = `\d+(?:\.\d+)?`
	header         = "flat  flat%   sum%        cum   cum%"
)

var (
	funcNameRegexpCompiled = regexp.MustCompile(funcNameRegexp)
	floatRegexpCompiled    = regexp.MustCompile(floatRegexp)
)

// ProfileFilter collects filters for extracting function names from a profile.
type ProfileFilter struct {
	// Include only lines starting with specified prefix
	FunctionPrefixes []string

	// Ignore all functions after the last dot even if includes the above prefix
	IgnoreFunctions []string
}

// GetAllFunctionNames extracts all function names from a profile text file, applying the given filter.
func GetAllFunctionNames(filePath string, filter ProfileFilter) (names []string, err error) {
	scanner, file, err := shared.GetScanner(filePath)
	if err != nil {
		return nil, fmt.Errorf("GetAllFunctionNames Failed: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("file close failed: %w", closeErr)
			}
		}
	}()

	ignoreSet := getFilterSets(filter.IgnoreFunctions)

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
			names = append(names, funcName)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading profile file: %w", err)
	}

	if !foundHeader {
		return nil, fmt.Errorf("profile file header not found")
	}

	return names, nil
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

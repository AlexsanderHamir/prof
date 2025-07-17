package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	funcNameRegexp = regexp.MustCompile(`\.([^.(]+)(?:\([^)]*\))?$`)
	floatRegexp    = regexp.MustCompile(`\d+(?:\.\d+)?`)
)

// ProfileFilter defines filters for extracting function names from a profile.
type ProfileFilter struct {
	FunctionPrefixes []string
	IgnoreFunctions  []string
}

// ExtractAllFunctionNames extracts all unique function names from a profile text file, applying the given filter.
func ExtractAllFunctionNames(profileTextFile string, filter ProfileFilter) ([]string, error) {
	file, err := os.Open(profileTextFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer file.Close()

	var functions []string
	ignoreSet := make(map[string]struct{})
	for _, f := range filter.IgnoreFunctions {
		ignoreSet[f] = struct{}{}
	}

	scanner := bufio.NewScanner(file)
	foundHeader := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, "flat  flat%   sum%        cum   cum%") {
			foundHeader = true
			continue
		}

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

	// Remove duplicates
	seen := make(map[string]struct{})
	var uniqueFunctions []string
	for _, f := range functions {
		if _, ok := seen[f]; !ok {
			seen[f] = struct{}{}
			uniqueFunctions = append(uniqueFunctions, f)
		}
	}

	return uniqueFunctions, nil
}

// ShouldKeepLine determines if a line from a profile should be kept based on profile values and ignore filters.
func ShouldKeepLine(line string, profileValues map[int]float64, ignoreFunctions, ignorePrefixes []string) bool {
	if line == "" {
		return false
	}

	parts := strings.Fields(line)
	if len(parts) < 6 {
		return false
	}

	ignoreSet := make(map[string]struct{})
	for _, f := range ignoreFunctions {
		ignoreSet[f] = struct{}{}
	}

	ignorePrefixSet := make(map[string]struct{})
	for _, p := range ignorePrefixes {
		ignorePrefixSet[p] = struct{}{}
	}

	// Filter by profile values
	if !filterByNumber(profileValues, parts) {
		return false
	}

	// Filter by ignore functions
	if !filterByIgnoreFunctions(ignoreSet, parts) {
		return false
	}

	// Filter by ignore prefixes
	return filterByIgnorePrefixes(ignorePrefixSet, parts)
}

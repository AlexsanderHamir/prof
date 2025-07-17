package parser

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	FlatIndex = iota
	FlatPercentIndex
	SumPercentIndex
	CumIndex
	CumPercentIndex
	MeasurementFieldCount
)

// filterByNumber returns true if all profile measurement values exceed their configured thresholds.
// It checks up to MeasurementFieldCount fields: flat, flat%, sum%, cum, cum%
func filterByNumber(thresholds map[int]float64, profileFields []string) bool {
	maxFields := min(MeasurementFieldCount, len(profileFields))

	for fieldIndex := range maxFields {
		threshold, hasThreshold := thresholds[fieldIndex]
		if !hasThreshold || threshold == 0.0 {
			continue
		}

		fieldValue, err := extractFloat(profileFields[fieldIndex])
		if err != nil {
			continue
		}

		if fieldValue <= threshold {
			return false
		}
	}
	return true
}

// filterByIgnoreFunctions returns false if the function is in the ignore set.
func filterByIgnoreFunctions(ignoreSet map[string]struct{}, parts []string) bool {
	if len(ignoreSet) == 0 {
		return true
	}

	fullFunctionName := cleanFunctionName(strings.Join(parts[5:], " "))
	_, ignored := ignoreSet[fullFunctionName]
	return !ignored
}

// filterByIgnorePrefixes returns false if the function name starts with any ignored prefix.
func filterByIgnorePrefixes(ignorePrefixSet map[string]struct{}, parts []string) bool {
	if len(ignorePrefixSet) == 0 {
		return true
	}

	fullFunctionName := strings.Join(parts[5:], " ")
	fullFunctionName = strings.ReplaceAll(fullFunctionName, " (inline)", "")
	fullFunctionName = strings.TrimSpace(fullFunctionName)

	for prefix := range ignorePrefixSet {
		if strings.HasPrefix(fullFunctionName, prefix) {
			return false
		}
	}

	return true
}

// cleanFunctionName returns the function name after the last dot, removing inline markers and trimming spaces.
func cleanFunctionName(s string) string {
	s = strings.ReplaceAll(s, " (inline)", "")
	s = strings.TrimSpace(s)

	// Get the part after the last dot
	lastDot := strings.LastIndex(s, ".")
	if lastDot != -1 && lastDot < len(s)-1 {
		return s[lastDot+1:]
	}
	return s
}

// extractFloat extracts the first float from a string.
func extractFloat(s string) (float64, error) {
	match := floatRegexp.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("no float found in '%s'", s)
	}
	return strconv.ParseFloat(match, 64)
}

// extractFunctionName extracts a function name from a line, applying prefix and ignore filters.
func extractFunctionName(line string, functionPrefixes []string, ignoreFunctions map[string]struct{}) string {
	parts := strings.Fields(line)
	if len(parts) < 6 {
		return ""
	}

	funcName := strings.Join(parts[5:], " ")

	// Check if function matches any prefix
	if len(functionPrefixes) > 0 {
		hasPrefix := false
		for _, prefix := range functionPrefixes {
			if strings.Contains(funcName, prefix) {
				hasPrefix = true
				break
			}
		}
		if !hasPrefix {
			return ""
		}
	}

	// Extract the actual function name (part after the last dot)
	matches := funcNameRegexp.FindStringSubmatch(funcName)
	if len(matches) < 2 {
		return ""
	}

	cleanName := strings.TrimSpace(strings.ReplaceAll(matches[1], " ", ""))
	if cleanName == "" {
		return ""
	}
	if _, ignored := ignoreFunctions[cleanName]; ignored {
		return ""
	}

	return cleanName
}

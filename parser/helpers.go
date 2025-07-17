package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// filterByNumber returns true if all relevant profile values in the line exceed the configured thresholds.
func filterByNumber(profileValues map[int]float64, parts []string) bool {
	for i := 0; i < 5 && i < len(parts); i++ {
		configValue, exists := profileValues[i]
		if !exists || configValue == 0.0 {
			continue
		}

		lineValue, err := extractFloat(parts[i])
		if err != nil {
			continue
		}

		if lineValue <= configValue {
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

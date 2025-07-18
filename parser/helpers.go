package parser

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	measurementFieldCount = 5
)

// filterByNumber returns true if all profile measurement values exceed their configured thresholds.
// It checks up to MeasurementFieldCount fields: flat, flat%, sum%, cum, cum%.
func filterByNumber(thresholds map[int]float64, profileFields []string) bool {
	maxFields := min(measurementFieldCount, len(profileFields))

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
	match := floatRegexpCompiled.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("no float found in '%s'", s)
	}
	return strconv.ParseFloat(match, 64)
}

func matchPrefix(funcName string, functionPrefixes []string) bool {
	var hasPrefix bool
	for _, prefix := range functionPrefixes {
		if strings.Contains(funcName, prefix) {
			hasPrefix = true
			break
		}
	}

	return hasPrefix
}

// extractFunctionName extracts a function name from a line, applying prefix and ignore filters.
func extractFunctionName(line string, functionPrefixes []string, ignoreFunctionSet map[string]struct{}) string {
	parts := strings.Fields(line)
	missingFields := len(parts) < profileLinelength
	if missingFields {
		return ""
	}

	funcName := strings.Join(parts[functionNameIndex:], " ")
	isPrefixConfigSet := len(functionPrefixes) > 0
	if isPrefixConfigSet && !matchPrefix(funcName, functionPrefixes) {
		return ""
	}

	matches := funcNameRegexpCompiled.FindStringSubmatch(funcName)
	// TODO: need more info
	if len(matches) < 2 {
		return ""
	}

	// TODO: need more info
	cleanName := strings.TrimSpace(strings.ReplaceAll(matches[1], " ", ""))
	if cleanName == "" {
		return ""
	}

	// TODO: need more info
	if _, ignored := ignoreFunctionSet[cleanName]; ignored {
		return ""
	}

	return cleanName
}

func getFilterSets(ignoreFunctions []string) map[string]struct{} {
	ignoreSet := make(map[string]struct{})
	for _, f := range ignoreFunctions {
		ignoreSet[f] = struct{}{}
	}

	return ignoreSet
}

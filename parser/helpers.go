package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	measurementFieldCount = 5 // total number of fields
	funcNameIndex         = 1 // The captured function name group
	minRequiredMatches    = 2 // Full match + at least one capture group
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
	missingFields := len(parts) < minProfileLinelength
	if missingFields {
		return ""
	}

	funcName := strings.Join(parts[functionNameIndex:], " ")
	isPrefixConfigSet := len(functionPrefixes) > 0
	if isPrefixConfigSet && !matchPrefix(funcName, functionPrefixes) {
		return ""
	}

	matches := funcNameRegexpCompiled.FindStringSubmatch(funcName)

	if len(matches) < minRequiredMatches {
		return ""
	}

	cleanName := strings.TrimSpace(strings.ReplaceAll(matches[funcNameIndex], " ", ""))
	if cleanName == "" {
		return ""
	}

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

var floatRegex = regexp.MustCompile(`^([+-]?\d*\.?\d+)`)

func convertToFloat(part string) (float64, error) {
	part = strings.TrimSpace(part)
	matches := floatRegex.FindStringSubmatch(part)

	if len(matches) < 2 {
		return 0, fmt.Errorf("failed to parse value '%s': no valid float found", part)
	}

	floatVal, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value '%s': %v", part, err)
	}

	return floatVal, nil
}

type ProfileFloats struct {
	Flat           float64
	FlatPercentage float64
	Sum            float64
	Cum            float64
	CumPercentage  float64
}

func getFloatsFromLineParts(lineParts []string) (*ProfileFloats, error) {
	flat, err := convertToFloat(lineParts[flatIndex])
	if err != nil {
		return nil, fmt.Errorf("flat conversion failed: %w", err)
	}

	flatPercentage, err := convertToFloat(lineParts[flatPercentageIndex])
	if err != nil {
		return nil, fmt.Errorf("flatPercentage conversion failed: %w", err)
	}

	sum, err := convertToFloat(lineParts[sumPercentageIndex])
	if err != nil {
		return nil, fmt.Errorf("sum conversion failed: %w", err)
	}

	cum, err := convertToFloat(lineParts[cumIndex])
	if err != nil {
		return nil, fmt.Errorf("cum conversion failed: %w", err)
	}

	cumPercentage, err := convertToFloat(lineParts[cumPercentageIndex])
	if err != nil {
		return nil, fmt.Errorf("cumPercentage conversion failed: %w", err)
	}

	return &ProfileFloats{
		Flat:           flat,
		FlatPercentage: flatPercentage,
		Sum:            sum,
		Cum:            cum,
		CumPercentage:  cumPercentage,
	}, nil
}

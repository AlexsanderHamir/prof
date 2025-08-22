package parser

import (
	"strings"
)

const (
	funcNameIndex      = 1 // The captured function name group
	minRequiredMatches = 2 // Full match + at least one capture group
)

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

// ExtractFunctionName extracts a function name from a line, applying prefix and ignore filters.
func ExtractFunctionName(line string, functionPrefixes []string, ignoreFunctionSet map[string]struct{}) string {
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

type ProfileFloats struct {
	Flat           float64
	FlatPercentage float64
	Sum            float64
	Cum            float64
	CumPercentage  float64
}

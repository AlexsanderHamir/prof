package parser

import "strings"

// simpleFunctionName returns the short name from a full symbol (e.g. "pkg.(*T).Method" → "Method").
func simpleFunctionName(fullPath string) string {
	parts := strings.Split(fullPath, ".")
	lastPart := parts[len(parts)-1]
	if idx := strings.Index(lastPart, "("); idx != -1 {
		lastPart = lastPart[:idx]
	}
	return lastPart
}

func matchPrefix(funcName string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.Contains(funcName, prefix) {
			return true
		}
	}
	return false
}

func ignoreSet(ignoreFunctions []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ignoreFunctions))
	for _, f := range ignoreFunctions {
		m[f] = struct{}{}
	}
	return m
}

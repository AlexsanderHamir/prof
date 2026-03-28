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

// packageNameFromSymbol derives a coarse module/package key from a full symbol name.
func packageNameFromSymbol(fullPath string) string {
	parts := strings.Split(fullPath, ".")
	if len(parts) < 2 {
		return ""
	}
	if !strings.Contains(parts[0], "/") && len(parts) >= 2 {
		if len(parts) >= 3 && strings.Contains(parts[1], "/") {
			return parts[0] + "." + parts[1]
		}
		return parts[0]
	}
	return parts[0]
}

func shortPackageLabel(fullPackageName string) string {
	parts := strings.Split(fullPackageName, ".")
	if strings.Contains(fullPackageName, "github.com") {
		return parts[len(parts)-1]
	}
	return fullPackageName
}

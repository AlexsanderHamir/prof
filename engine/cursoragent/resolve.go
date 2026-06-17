package cursoragent

// Adapted from T2A pkgs/agents/runner/adapterkit/probe.go (ResolveBinaryPath, FirstNonEmptyLine, TrimForLog).

import (
	"strings"

	"github.com/AlexsanderHamir/prof/engine/tooling"
)

// MergeBinaryPath returns the first non-empty trimmed value in order: flag, env, then empty string.
// Callers combine CLI flag and [EnvBinaryOverride] before [Options.BinaryPath].
func MergeBinaryPath(flagValue, envValue string) string {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v
	}
	if v := strings.TrimSpace(envValue); v != "" {
		return v
	}
	return ""
}

// ResolveBinaryPath returns the path exec would use after PATH lookup, or the trimmed input when lookup fails.
func ResolveBinaryPath(binaryPath string) string {
	p := strings.TrimSpace(binaryPath)
	if p == "" {
		return ""
	}
	if abs, err := tooling.LookPath(p); err == nil {
		return abs
	}
	return p
}

func firstNonEmptyLine(b []byte) string {
	for _, line := range strings.Split(string(b), "\n") {
		if v := strings.TrimSpace(line); v != "" {
			return v
		}
	}
	return ""
}

func trimForLog(b []byte, maxBytes int) string {
	if maxBytes <= 0 {
		maxBytes = 256
	}
	s := strings.TrimSpace(string(b))
	if len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "…"
}

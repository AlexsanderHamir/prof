package cli

import (
	"fmt"
	"strings"
)

// validateListArguments checks if the benchmarks and profiles arguments are valid lists.
func validateListArguments(benchmarks, profiles string) error {
	if strings.TrimSpace(benchmarks) == "[]" {
		return ErremptyBenchmarks
	}
	if strings.TrimSpace(profiles) == "[]" {
		return ErremptyProfiles
	}

	benchmarks = strings.TrimSpace(benchmarks)
	profiles = strings.TrimSpace(profiles)

	if !strings.HasPrefix(benchmarks, "[") || !strings.HasSuffix(benchmarks, "]") {
		return fmt.Errorf("benchmarks %w %s", Errbracket, benchmarks)
	}
	if !strings.HasPrefix(profiles, "[") || !strings.HasSuffix(profiles, "]") {
		return fmt.Errorf("profiles %w %s", Errbracket, profiles)
	}

	return nil
}

// parseListArgument parses a bracketed, comma-separated string into a slice of strings.
func parseListArgument(arg string) []string {
	arg = strings.Trim(arg, "[]")
	if arg == "" {
		return []string{}
	}

	parts := strings.Split(arg, ",")
	var result []string
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

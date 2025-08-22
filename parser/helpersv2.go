package parser

import (
	"fmt"
	"os"
	"sort"
	"strings"

	pprofprofile "github.com/google/pprof/profile"
)

// ProfileData contains the extracted flat and cumulative data from a pprof profile
type ProfileData struct {
	Flat  map[string]int64
	Cum   map[string]int64
	Total int64
	// Pre-calculated percentages for easy access
	FlatPercentages map[string]float64
	CumPercentages  map[string]float64
	SumPercentages  map[string]float64
	// Sorted entries for easy iteration
	SortedEntries []FuncEntry
}

// FuncEntry represents a function with its flat value, sorted by flat value (descending)
type FuncEntry struct {
	Name string
	Flat int64
}

// extractProfileData extracts flat and cumulative data from a pprof profile file
func extractProfileData(profilePath string) (*ProfileData, error) {
	// Open and parse profile file
	p, total, err := parseProfileFile(profilePath)
	if err != nil {
		return nil, err
	}

	// Extract flat and cumulative data
	flat, cum := extractFlatAndCumulativeData(p)

	// Calculate percentages and sort entries
	flatPercentages, cumPercentages, sumPercentages, sortedEntries := calculatePercentagesAndSort(flat, cum, total)

	return &ProfileData{
		Flat:            flat,
		Cum:             cum,
		Total:           total,
		FlatPercentages: flatPercentages,
		CumPercentages:  cumPercentages,
		SumPercentages:  sumPercentages,
		SortedEntries:   sortedEntries,
	}, nil
}

// parseProfileFile opens and parses a pprof profile file, returning the profile and total samples
func parseProfileFile(profilePath string) (*pprofprofile.Profile, int64, error) {
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	p, err := pprofprofile.Parse(f)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse pprof profile: %w", err)
	}

	// Calculate total samples
	var total int64
	for _, s := range p.Sample {
		total += s.Value[0]
	}

	return p, total, nil
}

// extractFlatAndCumulativeData processes profile samples to extract flat and cumulative function data
func extractFlatAndCumulativeData(p *pprofprofile.Profile) (map[string]int64, map[string]int64) {
	flat := make(map[string]int64)
	cum := make(map[string]int64)

	// Process each sample
	for _, s := range p.Sample {
		value := s.Value[0]
		extractSampleData(s, value, flat, cum)
	}

	return flat, cum
}

// extractSampleData processes a single sample to update flat and cumulative maps
func extractSampleData(s *pprofprofile.Sample, value int64, flat, cum map[string]int64) {
	// Cumulative: add to all stack frames
	seenFuncs := make(map[string]bool)
	for _, loc := range s.Location {
		for _, line := range loc.Line {
			if line.Function == nil {
				continue
			}
			fn := line.Function.Name
			if !seenFuncs[fn] {
				cum[fn] += value
				seenFuncs[fn] = true
			}
		}
	}

	// Flat: top of stack only
	if len(s.Location) > 0 {
		topLoc := s.Location[0]
		if len(topLoc.Line) > 0 && topLoc.Line[0].Function != nil {
			fn := topLoc.Line[0].Function.Name
			flat[fn] += value
		}
	}
}

// calculatePercentagesAndSort calculates all percentages and sorts entries by flat value
func calculatePercentagesAndSort(flat, cum map[string]int64, total int64) (map[string]float64, map[string]float64, map[string]float64, []FuncEntry) {
	flatPercentages := make(map[string]float64)
	cumPercentages := make(map[string]float64)
	sumPercentages := make(map[string]float64)

	// Sort by flat value (descending) for sum percentage calculation
	entries := createSortedEntries(flat)

	// Calculate percentages
	calculateAllPercentages(entries, cum, total, flatPercentages, cumPercentages, sumPercentages)

	return flatPercentages, cumPercentages, sumPercentages, entries
}

// createSortedEntries creates a sorted slice of function entries by flat value
func createSortedEntries(flat map[string]int64) []FuncEntry {
	var entries []FuncEntry
	for fn, flatVal := range flat {
		entries = append(entries, FuncEntry{
			Name: fn,
			Flat: flatVal,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Flat > entries[j].Flat
	})

	return entries
}

// calculateAllPercentages calculates flat, cumulative, and sum percentages for all functions
func calculateAllPercentages(entries []FuncEntry, cum map[string]int64, total int64, flatPercentages, cumPercentages, sumPercentages map[string]float64) {
	percentageMultiplier := 100.0
	var running float64

	for _, entry := range entries {
		fn := entry.Name
		flatVal := entry.Flat
		cumVal := cum[fn]

		flatPct := float64(flatVal) / float64(total) * percentageMultiplier
		cumPct := float64(cumVal) / float64(total) * percentageMultiplier
		running += flatPct

		flatPercentages[fn] = flatPct
		cumPercentages[fn] = cumPct
		sumPercentages[fn] = running
	}
}

// extractSimpleFunctionName extracts just the function name from a full function path
func extractSimpleFunctionName(fullPath string) string {
	// Handle cases like "github.com/user/pkg.(*Type).Method" => Method

	// Split by dots and get the last part
	parts := strings.Split(fullPath, ".")
	if len(parts) == 0 {
		return ""
	}

	lastPart := parts[len(parts)-1]

	// Handle method calls like "(*Type).Method"
	if strings.Contains(lastPart, ").") {
		methodParts := strings.Split(lastPart, ").")
		if len(methodParts) > 1 {
			return methodParts[1]
		}
	}

	// Handle generic types like "Type[Param].Method"
	if strings.Contains(lastPart, "].)") {
		methodParts := strings.Split(lastPart, "].)")
		if len(methodParts) > 1 {
			return methodParts[1]
		}
	}

	// Remove any trailing parentheses and parameters
	if idx := strings.Index(lastPart, "("); idx != -1 {
		lastPart = lastPart[:idx]
	}

	return lastPart
}

package parser

import (
	"fmt"
	"os"
	"sort"

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
	// Open profile file
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	// Parse pprof profile
	p, err := pprofprofile.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pprof profile: %w", err)
	}

	// Calculate total samples
	var total int64
	for _, s := range p.Sample {
		total += s.Value[0]
	}

	// Maps to store flat and cumulative values
	flat := make(map[string]int64)
	cum := make(map[string]int64)

	// Process each sample
	for _, s := range p.Sample {
		value := s.Value[0]

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

	// Calculate all percentages upfront
	flatPercentages := make(map[string]float64)
	cumPercentages := make(map[string]float64)
	sumPercentages := make(map[string]float64)

	// Sort by flat value (descending) for sum percentage calculation
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

	// Calculate percentages
	var running float64
	for _, entry := range entries {
		fn := entry.Name
		flatVal := entry.Flat
		cumVal := cum[fn]

		flatPct := float64(flatVal) / float64(total) * 100
		cumPct := float64(cumVal) / float64(total) * 100
		running += flatPct

		flatPercentages[fn] = flatPct
		cumPercentages[fn] = cumPct
		sumPercentages[fn] = running
	}

	return &ProfileData{
		Flat:            flat,
		Cum:             cum,
		Total:           total,
		FlatPercentages: flatPercentages,
		CumPercentages:  cumPercentages,
		SumPercentages:  sumPercentages,
		SortedEntries:   entries,
	}, nil
}

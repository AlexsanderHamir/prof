package parser

import (
	"sort"

	pprofprofile "github.com/google/pprof/profile"
)

// AggregateProfileData builds [ProfileData] from a validated profile at valueIndex in each sample's Value slice.
func AggregateProfileData(p *pprofprofile.Profile, valueIndex int) *ProfileData {
	flat, cum := flatAndCumulativeFromSamples(p, valueIndex)
	total := totalSampleValue(p, valueIndex)
	flatPct, cumPct, sumPct, sorted := percentagesAndSort(flat, cum, total)
	return &ProfileData{
		Flat:            flat,
		Cum:             cum,
		Total:           total,
		FlatPercentages: flatPct,
		CumPercentages:  cumPct,
		SumPercentages:  sumPct,
		SortedEntries:   sorted,
	}
}

func totalSampleValue(p *pprofprofile.Profile, valueIndex int) int64 {
	var total int64
	for _, s := range p.Sample {
		total += s.Value[valueIndex]
	}
	return total
}

func flatAndCumulativeFromSamples(p *pprofprofile.Profile, valueIndex int) (map[string]int64, map[string]int64) {
	flat := make(map[string]int64)
	cum := make(map[string]int64)
	for _, s := range p.Sample {
		accumulateSample(s, s.Value[valueIndex], flat, cum)
	}
	return flat, cum
}

func accumulateSample(s *pprofprofile.Sample, value int64, flat, cum map[string]int64) {
	seen := make(map[string]bool)
	for _, loc := range s.Location {
		for _, line := range loc.Line {
			if line.Function == nil {
				continue
			}
			fn := line.Function.Name
			if !seen[fn] {
				cum[fn] += value
				seen[fn] = true
			}
		}
	}
	if len(s.Location) == 0 {
		return
	}
	topLoc := s.Location[0]
	if len(topLoc.Line) > 0 && topLoc.Line[0].Function != nil {
		flat[topLoc.Line[0].Function.Name] += value
	}
}

func percentagesAndSort(flat, cum map[string]int64, total int64) (map[string]float64, map[string]float64, map[string]float64, []FuncEntry) {
	flatPct := make(map[string]float64)
	cumPct := make(map[string]float64)
	sumPct := make(map[string]float64)
	entries := sortedFuncEntries(flat)
	fillPercentages(entries, cum, total, flatPct, cumPct, sumPct)
	return flatPct, cumPct, sumPct, entries
}

func sortedFuncEntries(flat map[string]int64) []FuncEntry {
	entries := make([]FuncEntry, 0, len(flat))
	for fn, flatVal := range flat {
		entries = append(entries, FuncEntry{Name: fn, Flat: flatVal})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Flat > entries[j].Flat })
	return entries
}

func fillPercentages(entries []FuncEntry, cum map[string]int64, total int64, flatPct, cumPct, sumPct map[string]float64) {
	const pctScale = 100.0
	var running float64
	for _, entry := range entries {
		fn := entry.Name
		flatVal := entry.Flat
		cumVal := cum[fn]
		fp := float64(flatVal) / float64(total) * pctScale
		cp := float64(cumVal) / float64(total) * pctScale
		running += fp
		flatPct[fn] = fp
		cumPct[fn] = cp
		sumPct[fn] = running
	}
}

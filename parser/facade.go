package parser

import "github.com/AlexsanderHamir/prof/internal/config"

// GetFunctionListEntriesFromProfileData returns per-function list targets after the same
// filtering as [GetAllFunctionNamesFromProfileData].
func GetFunctionListEntriesFromProfileData(d *ProfileData, filter config.FunctionFilter) []FunctionListEntry {
	if d == nil {
		return nil
	}
	ign := ignoreSet(filter.IgnoreFunctions)
	var entries []FunctionListEntry
	for _, entry := range d.SortedEntries {
		fn := entry.Name
		short := simpleFunctionName(fn)
		if short == "" {
			continue
		}
		if _, skip := ign[short]; skip {
			continue
		}
		if len(filter.IncludePrefixes) > 0 && !matchPrefix(fn, filter.IncludePrefixes) {
			continue
		}
		entries = append(entries, FunctionListEntry{OutputStem: short, FullSymbol: fn})
	}
	return entries
}

// GetAllFunctionNamesFromProfileData applies filters to [ProfileData] the same way as [GetAllFunctionNamesV2].
func GetAllFunctionNamesFromProfileData(d *ProfileData, filter config.FunctionFilter) []string {
	if d == nil {
		return nil
	}
	entries := GetFunctionListEntriesFromProfileData(d, filter)
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.OutputStem
	}
	return names
}

// GetFunctionListEntriesV2 loads a profile path and returns [FunctionListEntry] values using filter rules.
func GetFunctionListEntriesV2(profilePath string, filter config.FunctionFilter) ([]FunctionListEntry, error) {
	d, err := profileDataFromPath(profilePath)
	if err != nil {
		return nil, err
	}
	return GetFunctionListEntriesFromProfileData(d, filter), nil
}

// GetAllFunctionNamesV2 extracts short function names from a profile path using filter rules.
func GetAllFunctionNamesV2(profilePath string, filter config.FunctionFilter) ([]string, error) {
	entries, err := GetFunctionListEntriesV2(profilePath, filter)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.OutputStem
	}
	return names, nil
}

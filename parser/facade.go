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

// OrganizeProfileByPackageFromProfileData builds the package-grouped markdown report from [ProfileData].
func OrganizeProfileByPackageFromProfileData(profileData *ProfileData, filter config.FunctionFilter) string {
	if profileData == nil {
		return ""
	}
	groups := make(map[string]*PackageGroup)
	ign := ignoreSet(filter.IgnoreFunctions)

	for _, entry := range profileData.SortedEntries {
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

		pkgName := packageNameFromSymbol(fn)
		if pkgName == "" {
			pkgName = "unknown"
		}
		if groups[pkgName] == nil {
			groups[pkgName] = &PackageGroup{
				Name:      pkgName,
				Functions: make([]*FunctionInfo, 0),
			}
		}
		info := &FunctionInfo{
			Name:           short,
			FullName:       fn,
			Flat:           float64(entry.Flat),
			FlatPercentage: profileData.FlatPercentages[fn],
			Cum:            float64(profileData.Cum[fn]),
			CumPercentage:  profileData.CumPercentages[fn],
			SumPercentage:  profileData.SumPercentages[fn],
		}
		g := groups[pkgName]
		g.Functions = append(g.Functions, info)
		g.TotalFlat += info.Flat
		g.TotalCum += info.Cum
	}

	totalFlat := float64(profileData.Total)
	for _, pkg := range groups {
		pkg.FlatPercentage = pkg.TotalFlat / totalFlat * 100
		pkg.CumPercentage = pkg.TotalCum / totalFlat * 100
	}
	return formatPackageReport(sortPackagesByFlatPercentage(groups))
}

// OrganizeProfileByPackageV2 loads a profile path and returns the package-grouped markdown report.
func OrganizeProfileByPackageV2(profilePath string, filter config.FunctionFilter) (string, error) {
	d, err := profileDataFromPath(profilePath)
	if err != nil {
		return "", err
	}
	return OrganizeProfileByPackageFromProfileData(d, filter), nil
}

// GetAllFunctionNames extracts function names from a profile path; equivalent to [GetAllFunctionNamesV2].
func GetAllFunctionNames(profilePath string, filter config.FunctionFilter) ([]string, error) {
	return GetAllFunctionNamesV2(profilePath, filter)
}

// OrganizeProfileByPackage loads a profile and builds the package-grouped report; equivalent to [OrganizeProfileByPackageV2].
func OrganizeProfileByPackage(profilePath string, filter config.FunctionFilter) (string, error) {
	return OrganizeProfileByPackageV2(profilePath, filter)
}

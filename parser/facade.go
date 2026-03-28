package parser

import "github.com/AlexsanderHamir/prof/internal"

// LineObjsFromProfileData builds line objects from aggregated profile data (e.g. [Pipeline.RunFromPath]).
func LineObjsFromProfileData(d *ProfileData) []*LineObj {
	if d == nil {
		return nil
	}
	out := make([]*LineObj, 0, len(d.SortedEntries))
	for _, entry := range d.SortedEntries {
		fn := entry.Name
		out = append(out, &LineObj{
			FnName:         fn,
			Flat:           float64(entry.Flat),
			FlatPercentage: d.FlatPercentages[fn],
			SumPercentage:  d.SumPercentages[fn],
			Cum:            float64(d.Cum[fn]),
			CumPercentage:  d.CumPercentages[fn],
		})
	}
	return out
}

// TurnLinesIntoObjectsV2 loads a profile path and returns line objects sorted by flat cost.
func TurnLinesIntoObjectsV2(profilePath string) ([]*LineObj, error) {
	d, err := profileDataFromPath(profilePath)
	if err != nil {
		return nil, err
	}
	return LineObjsFromProfileData(d), nil
}

// GetAllFunctionNamesFromProfileData applies filters to [ProfileData] the same way as [GetAllFunctionNamesV2].
func GetAllFunctionNamesFromProfileData(d *ProfileData, filter internal.FunctionFilter) []string {
	if d == nil {
		return nil
	}
	ign := ignoreSet(filter.IgnoreFunctions)
	var names []string
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
		names = append(names, short)
	}
	return names
}

// GetAllFunctionNamesV2 extracts short function names from a profile path using filter rules.
func GetAllFunctionNamesV2(profilePath string, filter internal.FunctionFilter) ([]string, error) {
	d, err := profileDataFromPath(profilePath)
	if err != nil {
		return nil, err
	}
	return GetAllFunctionNamesFromProfileData(d, filter), nil
}

// OrganizeProfileByPackageFromProfileData builds the package-grouped markdown report from [ProfileData].
func OrganizeProfileByPackageFromProfileData(profileData *ProfileData, filter internal.FunctionFilter) string {
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
func OrganizeProfileByPackageV2(profilePath string, filter internal.FunctionFilter) (string, error) {
	d, err := profileDataFromPath(profilePath)
	if err != nil {
		return "", err
	}
	return OrganizeProfileByPackageFromProfileData(d, filter), nil
}

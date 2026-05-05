package parser

// LineObj is one function row for tracker-style comparisons (flat/cum and percentages).
type LineObj struct {
	FnName         string
	Flat           float64
	FlatPercentage float64
	SumPercentage  float64
	Cum            float64
	CumPercentage  float64
}

// PackageGroup is a module/package bucket for grouped report output.
type PackageGroup struct {
	Name           string
	Functions      []*FunctionInfo
	TotalFlat      float64
	TotalCum       float64
	FlatPercentage float64
	CumPercentage  float64
}

// FunctionInfo is one function inside a [PackageGroup].
type FunctionInfo struct {
	Name           string
	FullName       string
	Flat           float64
	FlatPercentage float64
	Cum            float64
	CumPercentage  float64
	SumPercentage  float64
}

// ProfileData holds aggregated flat/cum counters and derived percentages after parsing a profile.
type ProfileData struct {
	Flat            map[string]int64
	Cum             map[string]int64
	Total           int64
	FlatPercentages map[string]float64
	CumPercentages  map[string]float64
	SumPercentages  map[string]float64
	SortedEntries   []FuncEntry
}

// FuncEntry is one symbol row sorted by flat cost (descending).
type FuncEntry struct {
	Name string
	Flat int64
}

// FunctionListEntry drives one per-function `go tool pprof -list` export.
// OutputStem is the basename for "<stem>.txt" (short name, same as ignore_functions).
// FullSymbol is the package-qualified name from the profile; it is used to build a
// robust -list pattern (see collector) because short names alone can fail to match
// pprof's graph node names.
type FunctionListEntry struct {
	OutputStem string
	FullSymbol string
}

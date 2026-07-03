package parser

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

package internal

// #1 Config Fields

// Config holds the main configuration for the prof tool.
type Config struct {
	FunctionFilter map[string]FunctionFilter `json:"function_collection_filter"`
}

// FunctionCollectionFilter defines filters for a specific benchmark,
// the filters are used when deciding which functions to collect
// code line level information for.
type FunctionFilter struct {
	// Prefixes: only collect functions starting with these prefixes
	// Example: []string{"github.com/myorg", "main."}
	IncludePrefixes []string `json:"include_prefixes,omitempty"`

	// IgnoreFunctions ignores the function name after the last dot.
	// Example: "Get,Set" excludes pool.Get() and cache.Set()
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`
}

// #2 - Function Arguments

type LineFilterArgs struct {
	ProfileFilters    map[int]float64
	IgnoreFunctionSet map[string]struct{}
	IgnorePrefixSet   map[string]struct{}
}

type CollectionArgs struct {
	Tag             string
	Profiles        []string
	BenchmarkName   string
	BenchmarkConfig FunctionFilter
}

type BenchArgs struct {
	Benchmarks []string
	Profiles   []string
	Count      int
	Tag        string
}

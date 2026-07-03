package config

// Config holds the main configuration for the prof tool.
type Config struct {
	Version    int        `json:"version"`
	Collection Collection `json:"collection,omitempty"`
}

// Collection holds function-extract filters for collect pipelines.
type Collection struct {
	Defaults       FunctionFilter            `json:"defaults,omitempty"`
	Benchmarks     map[string]FunctionFilter `json:"benchmarks,omitempty"`
	ManualProfiles map[string]FunctionFilter `json:"manual_profiles,omitempty"`
}

// FunctionFilter defines filters for collection (per-function extracts).
type FunctionFilter struct {
	IncludePrefixes []string `json:"include_prefixes,omitempty"`
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`
}

// CollectionArgs describes one benchmark collection run.
type CollectionArgs struct {
	Tag             string
	Profiles        []string
	BenchmarkName   string
	BenchmarkConfig FunctionFilter
}

// AutoArgs holds arguments for the auto-benchmark command.
type AutoArgs struct {
	Benchmarks []string
	Profiles   []string
	Count      int
	Tag        string
}

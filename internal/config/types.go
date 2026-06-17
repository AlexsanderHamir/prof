package config

// Config holds the main configuration for the prof tool.
type Config struct {
	Version    int        `json:"version"`
	Collection Collection `json:"collection,omitempty"`
	Track      Track      `json:"track,omitempty"`
}

// Collection holds function-extract filters for collect pipelines.
type Collection struct {
	Defaults       FunctionFilter            `json:"defaults,omitempty"`
	Benchmarks     map[string]FunctionFilter `json:"benchmarks,omitempty"`
	ManualProfiles map[string]FunctionFilter `json:"manual_profiles,omitempty"`
}

// FunctionFilter defines filters for collection (per-function extracts and grouped text).
type FunctionFilter struct {
	IncludePrefixes []string `json:"include_prefixes,omitempty"`
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`
}

// Track holds regression comparison and CI gate policy for prof track.
type Track struct {
	Defaults   TrackPolicy            `json:"defaults,omitempty"`
	Benchmarks map[string]TrackPolicy `json:"benchmarks,omitempty"`
}

// TrackPolicy defines track-time ignores and regression thresholds.
type TrackPolicy struct {
	IgnoreFunctions      []string `json:"ignore_functions,omitempty"`
	IgnorePrefixes       []string `json:"ignore_prefixes,omitempty"`
	MinChangePercent     float64  `json:"min_change_percent,omitempty"`
	MaxRegressionPercent float64  `json:"max_regression_percent,omitempty"`
	FailOnImprovement    bool     `json:"fail_on_improvement,omitempty"`
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

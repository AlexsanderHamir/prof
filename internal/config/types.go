package config

// Config holds the main configuration for the prof tool.
type Config struct {
	FunctionFilter map[string]FunctionFilter `json:"function_collection_filter"`
	CIConfig       *CIConfig                 `json:"ci_config,omitempty"`
}

// FunctionFilter defines filters for a specific benchmark when collecting line-level data.
type FunctionFilter struct {
	IncludePrefixes []string `json:"include_prefixes,omitempty"`
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`
}

// CIConfig holds CI/CD specific configuration for performance tracking.
type CIConfig struct {
	Global     *CITrackingConfig           `json:"global,omitempty"`
	Benchmarks map[string]CITrackingConfig `json:"benchmarks,omitempty"`
}

// CITrackingConfig defines CI/CD specific filtering for performance tracking.
type CITrackingConfig struct {
	IgnoreFunctions        []string `json:"ignore_functions,omitempty"`
	IgnorePrefixes         []string `json:"ignore_prefixes,omitempty"`
	MinChangeThreshold     float64  `json:"min_change_threshold,omitempty"`
	MaxRegressionThreshold float64  `json:"max_regression_threshold,omitempty"`
	FailOnImprovement      bool     `json:"fail_on_improvement,omitempty"`
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

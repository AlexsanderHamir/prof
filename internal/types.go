package internal

// #1 Config Fields

// Config holds the main configuration for the prof tool.
type Config struct {
	FunctionFilter map[string]FunctionFilter `json:"function_collection_filter"`
	// CI/CD specific configuration for performance tracking
	CIConfig *CIConfig `json:"ci_config,omitempty"`
}

// FunctionCollectionFilter defines filters for a specific benchmark,
// the filters are used when deciding which functions to collect
// code line level information for.
type FunctionFilter struct {
	// Prefixes: only collect functions starting with these prefixes
	// Example: []string{"github.com/example/GenPool"}
	IncludePrefixes []string `json:"include_prefixes,omitempty"`

	// IgnoreFunctions ignores the function name after the last dot.
	// Example: "Get,Set" excludes pool.Get() and cache.Set()
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`
}

// CIConfig holds CI/CD specific configuration for performance tracking
type CIConfig struct {
	// Global CI/CD settings that apply to all tracking operations
	Global *CITrackingConfig `json:"global,omitempty"`

	// Benchmark-specific CI/CD settings
	Benchmarks map[string]CITrackingConfig `json:"benchmarks,omitempty"`
}

// CITrackingConfig defines CI/CD specific filtering for performance tracking
type CITrackingConfig struct {
	// Functions to ignore during performance comparison (reduces noise)
	// These functions won't cause CI/CD failures even if they regress
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`

	// Function prefixes to ignore during performance comparison
	// Example: ["runtime.", "reflect."] ignores all runtime and reflect functions
	IgnorePrefixes []string `json:"ignore_prefixes,omitempty"`

	// Minimum change threshold for CI/CD failure
	// Only functions with changes >= this threshold will cause failures
	MinChangeThreshold float64 `json:"min_change_threshold,omitempty"`

	// Maximum acceptable regression percentage for CI/CD
	// Overrides command-line regression threshold if set
	MaxRegressionThreshold float64 `json:"max_regression_threshold,omitempty"`

	// Whether to fail on improvements (useful for detecting unexpected optimizations)
	FailOnImprovement bool `json:"fail_on_improvement,omitempty"`
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

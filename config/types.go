package config

// Config holds the main configuration for the prof tool.
type Config struct {
	FunctionFilter map[string]FunctionFilter `json:"function_collection_filter"`
	AIConfig       AIConfig                  `json:"ai_config"`
}

// ModelConfig holds the configuration for the AI model.
type ModelConfig struct {
	Model              string  `json:"model,omitempty"`
	MaxTokens          int     `json:"max_tokens,omitempty"`
	Temperature        float32 `json:"temperature,omitempty"`
	TopP               float32 `json:"top_p,omitempty"`
	PromptFileLocation string  `json:"prompt_file_location,omitempty"`
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

// AIConfig holds configuration for all AI-driven analyses.
type AIConfig struct {
	APIKey      string      `json:"api_key,omitempty"`
	BaseURL     string      `json:"base_url,omitempty"`
	ModelConfig ModelConfig `json:"model_config"`
	// AllBenchmarks decides wheter or not to analyze all benchmarks.
	AllBenchmarks bool `json:"all_benchmarks,omitempty"`

	// AllProfiles decides wheter or not to analyze all profiles.
	AllProfiles bool `json:"all_profiles,omitempty"`

	// SpecificBenchmarks must be set when AllBenchmarks is false.
	SpecificBenchmarks []string `json:"specific_benchmarks,omitempty"`

	// SpecificProfiles must be set when AllProfiles is false.
	SpecificProfiles []string `json:"specific_profiles,omitempty"`

	// ProfileFilter: filter profile data before sending to AI model
	// Only data above these thresholds will be included
	ProfileFilter *ProfileFilter `json:"profile_filter,omitempty"`
}

// ProfileFilter is responsible for removing excess information from being
// passed to the model.
type ProfileFilter struct {
	// Thresholds define the values above which nothing will be passed.
	// Example: any function with sum% > 80% won't be included.
	Thresholds FilterValues `json:"thresholds"`

	// IgnoreFunctions ignores the function name after the last dot.
	// // Example: "Get,Set" excludes pool.Get() and cache.Set()
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`

	// IgnoreFunctions ignores the function name after the last dot.
	// Example: "Get,Set" excludes pool.Get() and cache.Set()
	IgnorePrefixes []string `json:"ignore_prefixes,omitempty"`
}

// ProfileValues holds threshold values for filtering profile data.
// It will cap the profile data at the speicfied values.
// TODO: Better explanation and examples must be provided.
type FilterValues struct {
	Flat        float64 `json:"flat,omitempty"`
	FlatPercent float64 `json:"flat%,omitempty"`
	SumPercent  float64 `json:"sum%,omitempty"`
	Cum         float64 `json:"cum,omitempty"`
	CumPercent  float64 `json:"cum%,omitempty"`
}

type ConfigBuilder struct {
	config Config
}

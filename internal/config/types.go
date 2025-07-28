package config

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

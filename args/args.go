package args

import "github.com/AlexsanderHamir/prof/config"

type ModelCallRequiredArgs struct {
	SystemPrompt   string
	ProfileContent string
	BenchmarkName  string
	ProfileType    string
}

type LineFilterArgs struct {
	ProfileFilters    map[int]float64
	IgnoreFunctionSet map[string]struct{}
	IgnorePrefixSet   map[string]struct{}
}

type CollectionArgs struct {
	Tag             string
	Profiles        []string
	BenchmarkName   string
	BenchmarkConfig config.FunctionFilter
}

type BenchArgs struct {
	Benchmarks []string
	Profiles   []string
	Count      int
	Tag        string
}

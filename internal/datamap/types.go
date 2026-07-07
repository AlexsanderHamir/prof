package datamap

// SchemaVersion is the map.json schema version written by prof.
const SchemaVersion = 1

// Purpose values for artifact sections (stable enums for LLM consumers).
const (
	PurposeBenchmemResults       = "go_test_benchmem_results"
	PurposeRawPprofBinary        = "raw_pprof_binary"
	PurposeFlatCumulativeRanking = "flat_and_cumulative_ranking"
	PurposeCallerCalleeContext   = "caller_callee_context"
	PurposeLineLevelSource       = "line_level_source_extract"
	PurposeVisualCallGraph       = "visual_call_graph"
)

// BenchmarkMap is the root document written to data_mapping/<Benchmark>/map.json.
type BenchmarkMap struct {
	SchemaVersion      int                           `json:"schema_version"`
	Tag                string                        `json:"tag"`
	Benchmark          string                        `json:"benchmark"`
	Package            string                        `json:"package,omitempty"`
	RecommendedFlow    []string                      `json:"recommended_flow"`
	ReadingGuide       map[string]string             `json:"reading_guide"`
	ProfileCostColumns map[string]string             `json:"profile_cost_columns"`
	ProfileCostTriage  string                        `json:"profile_cost_triage"`
	Measurements       *MeasurementsSection          `json:"measurements,omitempty"`
	Profiles           map[string]ProfileRef         `json:"profiles"`
	Hotspots           map[string]HotspotSection     `json:"hotspots"`
	CallTrees          map[string]CallTreeSection    `json:"call_trees"`
	SourceLines        map[string]SourceLinesSection `json:"source_lines"`
	CallGraphs         map[string]CallGraphRef       `json:"call_graphs,omitempty"`
	Provenance         Provenance                    `json:"provenance"`
	Status             Status                        `json:"status"`
}

// MeasurementsSection points at go test bench output.
type MeasurementsSection struct {
	Path        string              `json:"path"`
	Purpose     string              `json:"purpose"`
	Description string              `json:"description"`
	Summary     *MeasurementSummary `json:"summary,omitempty"`
}

// MeasurementSummary holds parsed headline numbers from run.txt.
type MeasurementSummary struct {
	Count          int     `json:"count,omitempty"`
	NsPerOpMedian  int64   `json:"ns_per_op_median,omitempty"`
	BytesPerOp     int64   `json:"bytes_per_op,omitempty"`
	AllocsPerOp    int64   `json:"allocs_per_op,omitempty"`
	ElapsedSeconds float64 `json:"elapsed_seconds,omitempty"`
	Result         string  `json:"result,omitempty"`
}

// ProfileRef describes a raw pprof binary.
type ProfileRef struct {
	Path         string  `json:"path"`
	Purpose      string  `json:"purpose"`
	Description  string  `json:"description"`
	TotalSamples int64   `json:"total_samples,omitempty"`
	SampleUnit   string  `json:"sample_unit,omitempty"`
	OutputUnit   string  `json:"output_unit,omitempty"`
	TotalDisplay string  `json:"total_display,omitempty"`
	TotalSeconds float64 `json:"total_seconds,omitempty"`
}

// HotspotSection describes a pprof -top text artifact.
type HotspotSection struct {
	Path                string `json:"path"`
	Purpose             string `json:"purpose"`
	Description         string `json:"description"`
	Producer            string `json:"producer"`
	HotspotsMetricsNote string `json:"hotspots_metrics_note,omitempty"`
}

// CallTreeSection describes a pprof -tree text artifact.
type CallTreeSection struct {
	Path        string `json:"path"`
	Purpose     string `json:"purpose"`
	Description string `json:"description"`
	Producer    string `json:"producer"`
}

// SourceLinesSection indexes per-function -list extracts for one profile kind.
type SourceLinesSection struct {
	Dir         string                 `json:"dir"`
	PathPattern string                 `json:"path_pattern"`
	Purpose     string                 `json:"purpose"`
	Description string                 `json:"description"`
	Producer    string                 `json:"producer"`
	Functions   map[string]FunctionRef `json:"functions"`
}

// FunctionRef is one source_lines extract (path mapping only; metrics are in hotspots text).
type FunctionRef struct {
	Path       string `json:"path"`
	FullSymbol string `json:"full_symbol"`
	Status     string `json:"status"`
}

// CallGraphRef describes an optional PNG call graph.
type CallGraphRef struct {
	Path        string `json:"path"`
	Purpose     string `json:"purpose"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Reason      string `json:"reason,omitempty"`
}

// Provenance records how the map was produced.
type Provenance struct {
	Tag               string         `json:"tag"`
	CollectionMode    string         `json:"collection_mode"`
	BenchCount        int            `json:"bench_count,omitempty"`
	ProfilesRequested []string       `json:"profiles_requested"`
	Filter            FilterSnapshot `json:"filter"`
}

// FilterSnapshot mirrors prof.json function filter at collect time.
type FilterSnapshot struct {
	IncludePrefixes []string `json:"include_prefixes,omitempty"`
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`
}

// Status summarizes artifact availability.
type Status struct {
	BenchmarkRun string                       `json:"benchmark_run,omitempty"`
	Profiles     map[string]string            `json:"profiles"`
	Hotspots     map[string]string            `json:"hotspots"`
	CallTrees    map[string]string            `json:"call_trees"`
	CallGraphs   map[string]CallGraphStatus   `json:"call_graphs,omitempty"`
	SourceLines  map[string]SourceLinesStatus `json:"source_lines"`
}

// CallGraphStatus reports PNG availability for one profile.
type CallGraphStatus struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// SourceLinesStatus counts per-function list results.
type SourceLinesStatus struct {
	Collected int `json:"collected"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

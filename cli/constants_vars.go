package cli

var (
	// Root command flags.
	benchmarks []string
	profiles   []string
	tag        string
	count      int

	// Baseline is the baseline tag name for track and tools commands.
	Baseline string
	// Current is the current tag name for track and tools commands.
	Current             string
	benchmarkName       string
	profileType         string
	outputFormat        string
	failOnRegression    bool
	regressionThreshold float64

	// Profile organization flags (also used by manual collect for group-by-package).
	groupByPackage bool

	// auto Benchmark-only flags (see cmd_collect.go).
	lenientProfiles bool
	skipPNG         bool
)

const (
	tuiPageSize          = 20
	minTagsForComparison = 2

	// 3 occurrences requires a const
	baseTagFlag    = "base"
	currentTagFlag = "current"
	benchNameFlag  = "bench-name"
	tagFlag        = "tag"
)

package cli

var (
	// Root command flags.
	benchmarks []string
	profiles   []string
	tag        string
	count      int

	// Track command flags.
	Baseline            string
	Current             string
	benchmarkName       string
	profileType         string
	outputFormat        string
	failOnRegression    bool
	regressionThreshold float64

	// Profile organization flags.
	groupByPackage bool
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

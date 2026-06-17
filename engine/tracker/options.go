package tracker

// Options captures CLI / TUI inputs for track auto or manual flows.
type Options struct {
	Baseline            string
	Current             string
	BenchmarkName       string
	ProfileType         string
	OutputFormat        string
	UseThreshold        bool
	RegressionThreshold float64
	IsManual            bool
}

// Selections is an alias kept for internal compare helpers during migration.
type Selections = Options

package tracker

// Selections captures CLI / TUI inputs for track auto or manual flows.
type Selections struct {
	Baseline            string
	Current             string
	BenchmarkName       string
	ProfileType         string
	OutputFormat        string
	UseThreshold        bool
	RegressionThreshold float64
	IsManual            bool
}

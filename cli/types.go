package cli

// Arguments holds the CLI arguments for the prof tool.
type Arguments struct {
	Version        bool
	Command        string
	CreateTemplate bool
	OutputPath     string
	Benchmarks     string
	Profiles       string
	Tag            string
	Count          int

	// Performs analyzes on specified profile, according to specified configuration
	// and saves the results in a different file under the AI directory.
	GeneralAnalyze bool

	// Rewrites the profile file instead of saving an analysis in a different place,
	// useful for flagging requests.
	FlagProfiles bool
}

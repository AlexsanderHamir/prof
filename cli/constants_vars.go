package cli

// Shared flag names for track and tools commands.
const (
	tuiPageSize          = 20
	minTagsForComparison = 2

	baseTagFlag    = "base"
	currentTagFlag = "current"
	benchNameFlag  = "bench-name"
	tagFlag        = "tag"
)

// toolsFlags holds prof tools subcommand flags.
type toolsFlags struct {
	baseline      string
	current       string
	benchmarkName string
	tag           string
	profileType   string
}

var toolsGlobal toolsFlags

package internal

// The point of this file was to eliminate magic values on the codebase

// CLI subcommand names for collect and track flows.
const (
	AUTOCMD        = "auto"
	MANUALCMD      = "manual"
	TrackAutoCMD   = AUTOCMD
	TrackManualCMD = MANUALCMD
)

// Messages and labels for benchmark results and CI-style comparisons.
const (
	InfoCollectionSuccess = "All benchmarks and profile processing completed successfully!"
	IMPROVEMENT           = "IMPROVEMENT"
	REGRESSION            = "REGRESSION"
	STABLE                = "STABLE"
)

// Output directory names and file layout constants for bench artifacts.
const (
	MainDirOutput      = "bench"
	ProfileTextDir     = "text"
	ToolDir            = "tools"
	ProfileBinDir      = "bin"
	PermDir            = 0o755
	PermFile           = 0o644
	FunctionsDirSuffix = "_functions"
	ToolsResultsSuffix = "_results.txt"
	TextExtension      = "txt"
	ConfigFilename     = "config_template.json"
	GlobalSign         = "*"
	ExpectedTestSuffix = ".test"
)

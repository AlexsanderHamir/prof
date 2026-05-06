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
	// ProfileArtifactExtension is the file extension for pprof binaries under bench/<tag>/bin/ (and go test output before rename).
	ProfileArtifactExtension = "out"
	// BenchDescriptionFileName is the tag-level readme next to bin/ and text/.
	BenchDescriptionFileName = "description.txt"
	// BenchDescriptionPlaceholder is the initial body written for BenchDescriptionFileName until the user edits it.
	BenchDescriptionPlaceholder = "The explanation for this profilling session goes here"
)

// Go toolchain argv fragments (go test, go tool pprof).
const (
	GoBinaryName     = "go"
	GoTestSubcommand = "test"
	goToolLiteral    = "tool"
	goPprofLiteral   = "pprof"
)

// GoToolPprofPrefix returns the argv prefix {"go","tool","pprof"} for subprocess builders.
func GoToolPprofPrefix() []string {
	return []string{GoBinaryName, goToolLiteral, goPprofLiteral}
}

// External CLIs under prof tools; ToolNameBenchstat also names bench/tools/<name>/ for results.
const (
	ToolNameBenchstat   = "benchstat"
	ToolNameQcachegrind = "qcachegrind"
)

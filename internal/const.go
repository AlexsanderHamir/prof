package internal

// The point of this file was to eliminate magic values on the codebase

const (
	AUTOCMD        = "auto"
	MANUALCMD      = "manual"
	TrackAutoCMD   = AUTOCMD
	TrackManualCMD = MANUALCMD
)

const (
	InfoCollectionSuccess = "All benchmarks and profile processing completed successfully!"
	IMPROVEMENT           = "IMPROVEMENT"
	REGRESSION            = "REGRESSION"
	STABLE                = "STABLE"
)

const (
	MainDirOutput      = "bench"
	ProfileTextDir     = "text"
	ProfileBinDir      = "bin"
	PermDir            = 0o755
	PermFile           = 0o644
	FunctionsDirSuffix = "_functions"
	TextExtension      = "txt"
	ConfigFilename     = "config_template.json"
	GlobalSign         = "*"
	ExpectedTestSuffix = ".test"
)

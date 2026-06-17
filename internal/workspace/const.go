package workspace

// Path and permission constants for bench/<tag>/ layout.
const (
	MainDirOutput               = "bench"
	ProfileTextDir              = "text"
	ToolDir                     = "tools"
	ProfileBinDir               = "bin"
	PermDir                     = 0o755
	PermFile                    = 0o644
	FunctionsDirSuffix          = "_functions"
	ToolsResultsSuffix          = "_results.txt"
	TextExtension               = "txt"
	ExpectedTestSuffix          = ".test"
	ProfileArtifactExtension    = "out"
	BenchDescriptionFileName    = "description.txt"
	BenchDescriptionPlaceholder = "The explanation for this profilling session goes here"
	GoBinaryName                = "go"
	GoTestSubcommand            = "test"
	ToolNameBenchstat           = "benchstat"
	ToolNameQcachegrind         = "qcachegrind"
)

// InfoCollectionSuccess is logged when auto collection completes.
const InfoCollectionSuccess = "All benchmarks and profile processing completed successfully!"

package workspace

// Path and permission constants for bench/<tag>/ layout.
// Domains follow domain/<benchmark>/artifact (profile kind is a segment under source_lines/ and call_graphs/).
const (
	MainDirOutput            = "bench"
	ProfilesDir              = "profiles"
	MeasurementsDir          = "measurements"
	HotspotsDir              = "hotspots"
	SourceLinesDir           = "source_lines"
	CallGraphsDir            = "call_graphs"
	MeasurementRunFile       = "run.txt"
	TagNotesFileName         = "notes.txt"
	TagNotesPlaceholder      = "The explanation for this profiling session goes here"
	PermDir                  = 0o755
	PermFile                 = 0o644
	TextExtension            = "txt"
	ExpectedTestSuffix       = ".test"
	ProfileArtifactExtension = "out"
	GoBinaryName             = "go"
	GoTestSubcommand         = "test"
)

// InfoCollectionSuccess is logged when auto collection completes.
const InfoCollectionSuccess = "All benchmarks and profile processing completed successfully!"

package workspace

// Path and permission constants for .prof/<tag>/ layout.
// Domains follow domain/<benchmark>/artifact (profile kind is a segment under source_lines/ and call_graphs/).
const (
	MainDirOutput            = ".prof"
	ProfilesDir              = "profiles"
	MeasurementsDir          = "measurements"
	HotspotsDir              = "hotspots"
	CallTreesDir             = "call_trees"
	SourceLinesDir           = "source_lines"
	CallGraphsDir            = "call_graphs"
	MeasurementRunFile       = "run.txt"
	TagNotesFileName         = "notes.txt"
	TagNotesPlaceholder      = "The explanation for this profiling session goes here"
	PermDir                  = 0o755
	PermFile                 = 0o644
	TextExtension            = "txt"
	JSONExtension            = "json"
	ExpectedTestSuffix       = ".test"
	ProfileArtifactExtension = "out"
	GoBinaryName             = "go"
	GoTestSubcommand         = "test"
)

// InfoCollectionSuccess is logged when auto collection completes.
const InfoCollectionSuccess = "All benchmarks and profile processing completed successfully!"

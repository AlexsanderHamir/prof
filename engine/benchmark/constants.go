package benchmark

// SupportedProfiles lists profile kinds supported by the benchmark pipeline.
var SupportedProfiles = []string{"cpu", "memory", "mutex", "block"}

// ProfileFlags maps profile names to go test profiling flags.
var ProfileFlags = map[string]string{
	"cpu":    "-cpuprofile=cpu.out",
	"memory": "-memprofile=memory.out",
	"mutex":  "-mutexprofile=mutex.out",
	"block":  "-blockprofile=block.out",
}

// ExpectedFiles maps profile names to expected pprof output filenames.
var ExpectedFiles = map[string]string{
	"cpu":    "cpu.out",
	"memory": "memory.out",
	"mutex":  "mutex.out",
	"block":  "block.out",
}

const (
	binExtension           = "out"
	descriptionFileName    = "description.txt"
	moduleNotFoundMsg      = "go: cannot find main module"
	descriptionFileMessage = "The explanation for this profilling session goes here"

	// Minimum number of regex capture groups expected for benchmark function
	minCaptureGroups = 2
)

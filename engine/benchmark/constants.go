package benchmark

var SupportedProfiles = []string{"cpu", "memory", "mutex", "block"}

var ProfileFlags = map[string]string{
	"cpu":    "-cpuprofile=cpu.out",
	"memory": "-memprofile=memory.out",
	"mutex":  "-mutexprofile=mutex.out",
	"block":  "-blockprofile=block.out",
}

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
	waitForFiles           = 100
	descritpionFileMessage = "The explanation for this profilling session goes here"

	// Minimum number of regex capture groups expected for benchmark function
	minCaptureGroups = 2
)

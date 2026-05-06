package benchmark

import "github.com/AlexsanderHamir/prof/engine/tooling"

const (
	binExtension           = "out"
	descriptionFileName    = "description.txt"
	moduleNotFoundMsg      = "go: cannot find main module"
	descriptionFileMessage = "The explanation for this profilling session goes here"

	// Minimum number of regex capture groups expected for benchmark function
	minCaptureGroups = 2
)

// benchmarkCatalog is the default profile catalog for the benchmark pipeline.
var benchmarkCatalog = tooling.DefaultCatalog()

// SupportedProfiles lists profile kinds supported by the benchmark pipeline (declaration order).
var SupportedProfiles = benchmarkCatalog.ProfileIDsSorted()

// ProfileFlags maps profile names to go test profiling flags.
var ProfileFlags = buildProfileFlags(benchmarkCatalog)

// ExpectedFiles maps profile names to expected pprof output filenames in the package directory before moves.
var ExpectedFiles = buildExpectedFiles(benchmarkCatalog)

func buildProfileFlags(c *tooling.Catalog) map[string]string {
	m := make(map[string]string)
	for _, k := range c.ProfileKinds() {
		m[k.ID] = k.GoTestFlag
	}
	return m
}

func buildExpectedFiles(c *tooling.Catalog) map[string]string {
	m := make(map[string]string)
	for _, k := range c.ProfileKinds() {
		m[k.ID] = k.OutFileName
	}
	return m
}

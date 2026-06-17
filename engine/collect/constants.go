package collect

import "github.com/AlexsanderHamir/prof/engine/tooling"

const (
	moduleNotFoundMsg = "go: cannot find main module"
	minCaptureGroups  = 2
)

var profileCatalog = tooling.DefaultCatalog()

// ProfileFlags maps profile names to go test profiling flags.
var ProfileFlags = buildProfileFlags(profileCatalog)

// ExpectedFiles maps profile names to expected pprof output filenames before moves.
var ExpectedFiles = buildExpectedFiles(profileCatalog)

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

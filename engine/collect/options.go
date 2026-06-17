// Package collect runs prof auto (go test) and prof manual (ingest) pipelines under bench/<tag>/.
package collect

// AutoOptions configures RunAuto.
type AutoOptions struct {
	Benchmarks      []string
	Profiles        []string
	Tag             string
	Count           int
	GroupByPackage  bool
	LenientProfiles bool
	SkipPNG         bool
}

// ManualOptions configures RunManual.
type ManualOptions struct {
	Files          []string
	Tag            string
	GroupByPackage bool
}

// SupportedProfiles lists profile kinds for auto collection.
var SupportedProfiles = profileCatalog.ProfileIDsSorted()

// Package collect runs prof auto (go test) and prof manual (ingest) pipelines under bench/<tag>/.
package collect

// AutoOptions configures RunAuto.
type AutoOptions struct {
	Benchmarks []string
	Profiles   []string
	Tag        string
	Count      int
}

// ManualOptions configures RunManual.
type ManualOptions struct {
	Files []string
	Tag   string
}

// SupportedProfiles lists profile kinds for auto collection.
var SupportedProfiles = profileCatalog.ProfileIDsSorted()

// Package collect runs prof auto (go test) and prof manual (ingest) pipelines under bench/<tag>/.
// Output domains: profiles/, measurements/, hotspots/, source_lines/, call_graphs/.
package collect

// AutoOptions configures RunAuto.
type AutoOptions struct {
	Benchmarks             []string
	Profiles               []string
	Tag                    string
	Count                  int
	MissingConfigWarnShown bool // survey already printed config.MissingConfigUserWarning
}

// ManualOptions configures RunManual.
type ManualOptions struct {
	Files []string
	Tag   string
}

// SupportedProfiles lists profile kinds for auto collection.
var SupportedProfiles = profileCatalog.ProfileIDsSorted()

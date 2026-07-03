package app

// CollectAutoOptions describes a prof auto run.
type CollectAutoOptions struct {
	Benchmarks      []string
	Profiles        []string
	Tag             string
	Count           int
	GroupByPackage  bool
	LenientProfiles bool
	SkipPNG         bool
}

// CollectManualOptions describes a prof manual ingest run.
type CollectManualOptions struct {
	Files          []string
	Tag            string
	GroupByPackage bool
}

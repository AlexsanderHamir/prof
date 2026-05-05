package app

import "github.com/AlexsanderHamir/prof/engine/tracker"

// Benchmark runs the auto benchmark pipeline and discovers benchmarks in the module.
type Benchmark interface {
	RunBenchmarks(benchmarks, profiles []string, tag string, count int, groupByPackage bool, lenientProfiles bool, skipPNG bool) error
	DiscoverBenchmarks(scope string) ([]string, error)
	SupportedProfiles() []string
}

// Collector organizes profile inputs under a tag (manual prof flow).
type Collector interface {
	RunCollector(files []string, tag string, groupByPackage bool) error
}

// Tracker compares baseline vs current profiles.
type Tracker interface {
	RunTrackAuto(selections *tracker.Selections) error
	RunTrackManual(selections *tracker.Selections) error
}

// Tools runs optional post-processing commands on collected data.
type Tools interface {
	RunBenchStats(baseTag, currentTag, benchName string) error
	RunQcacheGrind(tag, benchName, profile string) error
}

// Setup generates project scaffolding such as the config template.
type Setup interface {
	CreateTemplate() error
}

// Services is the composition root: inject alternate implementations for tests or custom backends.
type Services struct {
	Benchmark Benchmark
	Collector Collector
	Tracker   Tracker
	Tools     Tools
	Setup     Setup
}

// WithDefaults returns a copy of s with any nil fields replaced by default engine implementations.
func (s *Services) WithDefaults() *Services {
	if s == nil {
		return Default()
	}
	out := *s
	if out.Benchmark == nil {
		out.Benchmark = defaultBenchmark{}
	}
	if out.Collector == nil {
		out.Collector = defaultCollector{}
	}
	if out.Tracker == nil {
		out.Tracker = defaultTracker{}
	}
	if out.Tools == nil {
		out.Tools = defaultTools{}
	}
	if out.Setup == nil {
		out.Setup = defaultSetup{}
	}
	return &out
}

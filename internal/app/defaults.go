package app

import (
	"github.com/AlexsanderHamir/prof/engine/benchmark"
	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/engine/tools/benchstats"
	"github.com/AlexsanderHamir/prof/engine/tools/qcachegrind"
	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal"
)

// Default returns stock services (production wiring).
func Default() *Services {
	return &Services{
		Benchmark: defaultBenchmark{},
		Collector: defaultCollector{},
		Tracker:   defaultTracker{},
		Tools:     defaultTools{},
		Setup:     defaultSetup{},
	}
}

type defaultBenchmark struct{}

func (defaultBenchmark) RunBenchmarks(benchmarks, profiles []string, tag string, count int, groupByPackage bool, lenientProfiles bool, skipPNG bool) error {
	return benchmark.RunBenchmarks(benchmarks, profiles, tag, count, groupByPackage, lenientProfiles, skipPNG)
}

func (defaultBenchmark) DiscoverBenchmarks(scope string) ([]string, error) {
	return benchmark.DiscoverBenchmarks(scope)
}

func (defaultBenchmark) SupportedProfiles() []string {
	return benchmark.SupportedProfiles
}

type defaultCollector struct{}

func (defaultCollector) RunCollector(files []string, tag string, groupByPackage bool) error {
	return collector.RunCollector(files, tag, groupByPackage)
}

type defaultTracker struct{}

func (defaultTracker) RunTrackAuto(selections *tracker.Selections) error {
	return tracker.RunTrackAuto(selections)
}

func (defaultTracker) RunTrackManual(selections *tracker.Selections) error {
	return tracker.RunTrackManual(selections)
}

type defaultTools struct{}

func (defaultTools) RunBenchStats(baseTag, currentTag, benchName string) error {
	return benchstats.RunBenchStats(baseTag, currentTag, benchName)
}

func (defaultTools) RunQcacheGrind(tag, benchName, profile string) error {
	return qcachegrind.RunQcacheGrind(tag, benchName, profile)
}

type defaultSetup struct{}

func (defaultSetup) CreateTemplate() error {
	return internal.CreateTemplate()
}

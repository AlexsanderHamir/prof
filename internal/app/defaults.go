package app

import (
	"github.com/AlexsanderHamir/prof/engine/benchmark"
	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/engine/tools/benchstats"
	"github.com/AlexsanderHamir/prof/engine/tools/qcachegrind"
	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal"
)

// Default returns stock services (production wiring).
func Default() *Services {
	r := tooling.NewExecRunner()
	return &Services{
		Runner:    r,
		Benchmark: defaultBenchmark{runner: r},
		Collector: defaultCollector{runner: r},
		Tracker:   defaultTracker{},
		Tools:     defaultTools{runner: r},
		Setup:     defaultSetup{},
	}
}

type defaultBenchmark struct {
	runner tooling.Runner
}

func (d defaultBenchmark) RunBenchmarks(benchmarks, profiles []string, tag string, count int, groupByPackage bool, lenientProfiles bool, skipPNG bool) error {
	return benchmark.RunBenchmarks(d.runner, benchmarks, profiles, tag, count, groupByPackage, lenientProfiles, skipPNG)
}

func (defaultBenchmark) DiscoverBenchmarks(scope string) ([]string, error) {
	return benchmark.DiscoverBenchmarks(scope)
}

func (defaultBenchmark) SupportedProfiles() []string {
	return benchmark.SupportedProfiles
}

type defaultCollector struct {
	runner tooling.Runner
}

func (d defaultCollector) RunCollector(files []string, tag string, groupByPackage bool) error {
	return collector.RunCollector(d.runner, files, tag, groupByPackage)
}

type defaultTracker struct{}

func (defaultTracker) RunTrackAuto(selections *tracker.Selections) error {
	return tracker.RunTrackAuto(selections)
}

func (defaultTracker) RunTrackManual(selections *tracker.Selections) error {
	return tracker.RunTrackManual(selections)
}

type defaultTools struct {
	runner tooling.Runner
}

func (d defaultTools) RunBenchStats(baseTag, currentTag, benchName string) error {
	return benchstats.RunBenchStats(d.runner, baseTag, currentTag, benchName)
}

func (d defaultTools) RunQcacheGrind(tag, benchName, profile string) error {
	return qcachegrind.RunQcacheGrind(d.runner, tag, benchName, profile)
}

type defaultSetup struct{}

func (defaultSetup) CreateTemplate() error {
	return internal.CreateTemplate()
}

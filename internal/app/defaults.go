package app

import (
	"context"

	"github.com/AlexsanderHamir/prof/engine/collect"
	"github.com/AlexsanderHamir/prof/engine/cursoragent"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/engine/tools/benchstats"
	"github.com/AlexsanderHamir/prof/engine/tools/qcachegrind"
	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal/config"
)

// Default returns stock services (production wiring).
func Default() *Services {
	r := tooling.NewExecRunner()
	return &Services{
		Runner:  r,
		Collect: defaultCollect{runner: r},
		Tracker: defaultTracker{},
		Tools:   defaultTools{runner: r},
		Agent:   defaultAgent{},
		Setup:   defaultSetup{},
	}
}

type defaultCollect struct {
	runner tooling.Runner
}

func (d defaultCollect) RunAuto(opts CollectAutoOptions) error {
	return collect.RunAuto(d.runner, collect.AutoOptions(opts))
}

func (d defaultCollect) RunManual(opts CollectManualOptions) error {
	return collect.RunManual(d.runner, collect.ManualOptions(opts))
}

func (d defaultCollect) DiscoverBenchmarks(scope string) ([]string, error) {
	return collect.DiscoverBenchmarks(scope)
}

func (d defaultCollect) SupportedProfiles() []string {
	return collect.SupportedProfiles
}

type defaultTracker struct{}

func (defaultTracker) RunTrackAuto(opts TrackOptions) error {
	return tracker.RunTrackAuto(tracker.Options(opts))
}

func (defaultTracker) RunTrackManual(opts TrackOptions) error {
	return tracker.RunTrackManual(tracker.Options(opts))
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

type defaultAgent struct{}

func (defaultAgent) Run(ctx context.Context, req cursoragent.RunRequest, opts cursoragent.Options) (cursoragent.RunResult, error) {
	return cursoragent.NewClient(opts).Run(ctx, req)
}

type defaultSetup struct{}

func (defaultSetup) CreateTemplate() error {
	return config.CreateTemplate()
}

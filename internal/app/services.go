package app

import (
	"context"

	"github.com/AlexsanderHamir/prof/engine/cursoragent"
	"github.com/AlexsanderHamir/prof/engine/tooling"
)

// Collect runs auto and manual profile collection pipelines.
type Collect interface {
	RunAuto(opts CollectAutoOptions) error
	RunManual(opts CollectManualOptions) error
	DiscoverBenchmarks(scope string) ([]string, error)
	SupportedProfiles() []string
}

// Tracker compares baseline vs current profiles.
type Tracker interface {
	RunTrackAuto(opts TrackOptions) error
	RunTrackManual(opts TrackOptions) error
}

// Tools runs optional post-processing commands on collected data.
type Tools interface {
	RunBenchStats(baseTag, currentTag, benchName string) error
	RunQcacheGrind(tag, benchName, profile string) error
}

// Agent runs the cursor-agent integration when configured.
type Agent interface {
	Run(ctx context.Context, req cursoragent.RunRequest, opts cursoragent.Options) (cursoragent.RunResult, error)
}

// Setup generates project scaffolding such as the config template.
type Setup interface {
	CreateTemplate() error
}

// Services is the composition root: inject alternate implementations for tests or custom backends.
type Services struct {
	Runner  tooling.Runner
	Collect Collect
	Tracker Tracker
	Tools   Tools
	Agent   Agent
	Setup   Setup
}

// WithDefaults returns a copy of s with any nil fields replaced by default engine implementations.
func (s *Services) WithDefaults() *Services {
	if s == nil {
		return Default()
	}
	out := *s
	if out.Runner == nil {
		out.Runner = tooling.NewExecRunner()
	}
	if out.Collect == nil {
		out.Collect = defaultCollect{runner: out.Runner}
	}
	if out.Tracker == nil {
		out.Tracker = defaultTracker{}
	}
	if out.Tools == nil {
		out.Tools = defaultTools{runner: out.Runner}
	}
	if out.Agent == nil {
		out.Agent = defaultAgent{}
	}
	if out.Setup == nil {
		out.Setup = defaultSetup{}
	}
	return &out
}

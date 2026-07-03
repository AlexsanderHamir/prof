package app

import (
	"context"

	"github.com/AlexsanderHamir/prof/engine/cursoragent"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
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

// Agent runs the cursor-agent integration when configured.
type Agent interface {
	Run(ctx context.Context, req cursoragent.RunRequest, opts cursoragent.Options) (cursoragent.RunResult, error)
}

// Setup generates project scaffolding such as the config file.
type Setup interface {
	CreateTemplate() error
}

// Config loads and saves prof.json beside go.mod.
type Config interface {
	Load() (*config.Config, error)
	Save(*config.Config) error
	CreateDefaultFile() error
	Path() (string, error)
}

// Services is the composition root: inject alternate implementations for tests or custom backends.
type Services struct {
	Runner  tooling.Runner
	Collect Collect
	Tracker Tracker
	Agent   Agent
	Setup   Setup
	Config  Config
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
	if out.Agent == nil {
		out.Agent = defaultAgent{}
	}
	if out.Setup == nil {
		out.Setup = defaultSetup{}
	}
	if out.Config == nil {
		out.Config = defaultConfig{}
	}
	return &out
}

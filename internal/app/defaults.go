package app

import (
	"context"

	"github.com/AlexsanderHamir/prof/engine/collect"
	"github.com/AlexsanderHamir/prof/engine/cursoragent"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
)

// Default returns stock services (production wiring).
func Default() *Services {
	r := tooling.NewExecRunner()
	return &Services{
		Runner:  r,
		Collect: defaultCollect{runner: r},
		Agent:   defaultAgent{},
		Config:  defaultConfig{},
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

type defaultAgent struct{}

func (defaultAgent) Run(ctx context.Context, req cursoragent.RunRequest, opts cursoragent.Options) (cursoragent.RunResult, error) {
	return cursoragent.NewClient(opts).Run(ctx, req)
}

type defaultConfig struct{}

func (defaultConfig) Load() (*config.Config, error) {
	return config.Load()
}

func (defaultConfig) Save(cfg *config.Config) error {
	return config.Save(cfg)
}

func (defaultConfig) CreateDefaultFile() error {
	return config.CreateDefaultFile()
}

func (defaultConfig) Path() (string, error) {
	return config.Path(config.Filename)
}

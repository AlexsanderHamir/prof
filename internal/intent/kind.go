package intent

import (
	"github.com/AlexsanderHamir/prof/internal/app"
)

// Kind identifies a supported translation workflow for interactive flows.
type Kind string

const (
	// KindCollect runs benchmarks and collects profiles under .prof/<tag>/.
	KindCollect Kind = "collect"
	// KindConfigCreate writes the default prof.json beside go.mod.
	KindConfigCreate Kind = "config_create"
)

// KindDescriptor pairs a Kind with a one-line description for listings and tests.
type KindDescriptor struct {
	K           Kind
	Description string
}

// AllKinds returns every supported workflow in stable order.
func AllKinds() []KindDescriptor {
	return []KindDescriptor{
		{KindCollect, "Collect benchmark profiles (Benchmark.RunBenchmarks)"},
		{KindConfigCreate, "Create default prof.json (Config.CreateDefaultFile)"},
	}
}

// Executable is a validated intent that can run against app.Services.
type Executable interface {
	Kind() Kind
	Validate() error
	Run(svc *app.Services) error
}

// RunValidated runs e after Validate; use from CLI glue when you already built an intent.
func RunValidated(e Executable, svc *app.Services) error {
	if err := e.Validate(); err != nil {
		return err
	}
	return e.Run(svc)
}

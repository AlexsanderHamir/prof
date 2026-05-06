package intent

import (
	"github.com/AlexsanderHamir/prof/internal/app"
)

// Kind identifies a supported translation workflow for interactive flows.
type Kind string

const (
	// KindCollect runs benchmarks and collects profiles under bench/<tag>/.
	KindCollect Kind = "collect"
	// KindCompare compares two tags using the bench/ tree (track auto).
	KindCompare Kind = "compare"
	// KindSetup writes the config template beside go.mod.
	KindSetup Kind = "setup"
	// KindToolsBenchstat runs benchstat between two tags for one benchmark.
	KindToolsBenchstat Kind = "tools_benchstat"
	// KindToolsQcachegrind opens a binary profile in qcachegrind.
	KindToolsQcachegrind Kind = "tools_qcachegrind"
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
		{KindCompare, "Compare two tagged runs (Tracker.RunTrackAuto)"},
		{KindSetup, "Create configuration template (Setup.CreateTemplate)"},
		{KindToolsBenchstat, "Benchstat between two tags (Tools.RunBenchStats)"},
		{KindToolsQcachegrind, "Open profile in qcachegrind (Tools.RunQcacheGrind)"},
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

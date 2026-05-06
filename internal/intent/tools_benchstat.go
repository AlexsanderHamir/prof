package intent

import (
	"errors"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/app"
)

// ToolsBenchstatIntent runs benchstat for one benchmark between two tags.
type ToolsBenchstatIntent struct {
	BaseTag    string
	CurrentTag string
	BenchName  string
}

// Kind implements [Executable].
func (i *ToolsBenchstatIntent) Kind() Kind { return KindToolsBenchstat }

// Validate implements [Executable].
func (i *ToolsBenchstatIntent) Validate() error {
	if strings.TrimSpace(i.BaseTag) == "" {
		return errors.New("tools benchstat intent: baseline tag is required")
	}
	if strings.TrimSpace(i.CurrentTag) == "" {
		return errors.New("tools benchstat intent: current tag is required")
	}
	if strings.TrimSpace(i.BenchName) == "" {
		return errors.New("tools benchstat intent: benchmark name is required")
	}
	if i.BaseTag == i.CurrentTag {
		return errors.New("tools benchstat intent: baseline and current tags must differ")
	}
	return nil
}

// Run implements [Executable].
func (i *ToolsBenchstatIntent) Run(svc *app.Services) error {
	return svc.Tools.RunBenchStats(i.BaseTag, i.CurrentTag, i.BenchName)
}

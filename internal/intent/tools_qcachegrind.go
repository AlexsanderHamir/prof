package intent

import (
	"errors"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/app"
)

// ToolsQcachegrindIntent opens a collected binary profile in qcachegrind.
type ToolsQcachegrindIntent struct {
	Tag         string
	BenchName   string
	ProfileType string
}

// Kind implements [Executable].
func (i *ToolsQcachegrindIntent) Kind() Kind { return KindToolsQcachegrind }

// Validate implements [Executable].
func (i *ToolsQcachegrindIntent) Validate() error {
	if strings.TrimSpace(i.Tag) == "" {
		return errors.New("tools qcachegrind intent: tag is required")
	}
	if strings.TrimSpace(i.BenchName) == "" {
		return errors.New("tools qcachegrind intent: benchmark name is required")
	}
	if strings.TrimSpace(i.ProfileType) == "" {
		return errors.New("tools qcachegrind intent: profile type is required")
	}
	return nil
}

// Run implements [Executable].
func (i *ToolsQcachegrindIntent) Run(svc *app.Services) error {
	return svc.Tools.RunQcacheGrind(i.Tag, i.BenchName, i.ProfileType)
}

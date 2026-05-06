package intent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal/app"
)

var compareValidFormats = map[string]bool{
	"summary":       true,
	"detailed":      true,
	"summary-html":  true,
	"detailed-html": true,
	"summary-json":  true,
	"detailed-json": true,
}

// CompareIntent wraps tag-based track auto selections (prof track auto, prof ui compare).
type CompareIntent struct {
	Selections *tracker.Selections
}

// Kind implements [Executable].
func (i *CompareIntent) Kind() Kind { return KindCompare }

// Validate checks fields required for RunTrackAuto.
func (i *CompareIntent) Validate() error {
	if i.Selections == nil {
		return errors.New("compare intent: selections is nil")
	}
	s := i.Selections
	if s.IsManual {
		return errors.New("compare intent: IsManual must be false for tag-based compare (use prof track manual for file paths)")
	}
	if strings.TrimSpace(s.Baseline) == "" {
		return errors.New("compare intent: baseline tag is required")
	}
	if strings.TrimSpace(s.Current) == "" {
		return errors.New("compare intent: current tag is required")
	}
	if strings.TrimSpace(s.BenchmarkName) == "" {
		return errors.New("compare intent: benchmark name is required")
	}
	if strings.TrimSpace(s.ProfileType) == "" {
		return errors.New("compare intent: profile type is required")
	}
	if strings.TrimSpace(s.OutputFormat) == "" {
		return errors.New("compare intent: output format is required")
	}
	if !compareValidFormats[s.OutputFormat] {
		return fmt.Errorf("compare intent: invalid output format %q", s.OutputFormat)
	}
	if s.UseThreshold && s.RegressionThreshold <= 0 {
		return errors.New("compare intent: regression threshold must be positive when fail-on-regression is enabled")
	}
	return nil
}

// Run implements [Executable].
func (i *CompareIntent) Run(svc *app.Services) error {
	return svc.Tracker.RunTrackAuto(i.Selections)
}

package intent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/app"
)

// CompareIntent wraps tag-based track auto selections (prof track auto, prof ui compare).
type CompareIntent struct {
	Options app.TrackOptions
}

// Kind implements [Executable].
func (i *CompareIntent) Kind() Kind { return KindCompare }

// Validate checks fields required for RunTrackAuto.
func (i *CompareIntent) Validate() error {
	s := i.Options
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
	if !app.ValidTrackOutputFormat(s.OutputFormat) {
		return fmt.Errorf("compare intent: invalid output format %q", s.OutputFormat)
	}
	if s.UseThreshold && s.RegressionThreshold <= 0 {
		return errors.New("compare intent: regression threshold must be positive when fail-on-regression is enabled")
	}
	return nil
}

// Run implements [Executable].
func (i *CompareIntent) Run(svc *app.Services) error {
	return svc.Tracker.RunTrackAuto(i.Options)
}

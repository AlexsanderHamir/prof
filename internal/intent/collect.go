package intent

import (
	"errors"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/app"
)

// CollectIntent mirrors prof auto / prof tui collect → Collect.RunAuto.
type CollectIntent struct {
	Benchmarks      []string
	Profiles        []string
	Tag             string
	Count           int
	LenientProfiles bool
	SkipPNG         bool
}

// Kind implements [Executable].
func (i *CollectIntent) Kind() Kind { return KindCollect }

// Normalize trims whitespace on the tag and drops empty benchmark/profile entries.
func (i *CollectIntent) Normalize() {
	i.Tag = strings.TrimSpace(i.Tag)
	i.Benchmarks = nonEmptyStrings(i.Benchmarks)
	i.Profiles = nonEmptyStrings(i.Profiles)
}

// Validate checks required fields for a collect run.
func (i *CollectIntent) Validate() error {
	if len(i.Benchmarks) == 0 {
		return errors.New("collect intent: at least one benchmark is required")
	}
	if len(i.Profiles) == 0 {
		return errors.New("collect intent: at least one profile type is required")
	}
	if i.Tag == "" {
		return errors.New("collect intent: tag is required")
	}
	if i.Count < 1 {
		return errors.New("collect intent: count must be at least 1")
	}
	return nil
}

// Run implements [Executable].
func (i *CollectIntent) Run(svc *app.Services) error {
	return svc.Collect.RunAuto(app.CollectAutoOptions{
		Benchmarks:      i.Benchmarks,
		Profiles:        i.Profiles,
		Tag:             i.Tag,
		Count:           i.Count,
		LenientProfiles: i.LenientProfiles,
		SkipPNG:         i.SkipPNG,
	})
}

func nonEmptyStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal/app"
)

type fakeTracker struct {
	last *tracker.Selections
	err  error
}

func (f *fakeTracker) RunTrackAuto(sel *tracker.Selections) error {
	f.last = sel
	return f.err
}

func (f *fakeTracker) RunTrackManual(*tracker.Selections) error {
	return errors.New("unexpected RunTrackManual")
}

func TestCompareIntent_Validate(t *testing.T) {
	t.Parallel()
	valid := &tracker.Selections{
		Baseline: "a", Current: "b", BenchmarkName: "Bench", ProfileType: "cpu",
		OutputFormat: "detailed", UseThreshold: false,
	}
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		i := &CompareIntent{Selections: valid}
		if err := i.Validate(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("nil selections", func(t *testing.T) {
		t.Parallel()
		i := &CompareIntent{}
		if err := i.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("manual not allowed", func(t *testing.T) {
		t.Parallel()
		s := *valid
		s.IsManual = true
		i := &CompareIntent{Selections: &s}
		if err := i.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()
		s := *valid
		s.OutputFormat = "nope"
		i := &CompareIntent{Selections: &s}
		if err := i.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("threshold without value", func(t *testing.T) {
		t.Parallel()
		s := *valid
		s.UseThreshold = true
		s.RegressionThreshold = 0
		i := &CompareIntent{Selections: &s}
		if err := i.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCompareIntent_Run(t *testing.T) {
	t.Parallel()
	ft := &fakeTracker{}
	svc := &app.Services{Tracker: ft}
	sel := &tracker.Selections{
		Baseline: "base", Current: "cur", BenchmarkName: "B", ProfileType: "cpu",
		OutputFormat: "summary",
	}
	i := &CompareIntent{Selections: sel}
	if err := RunValidated(i, svc); err != nil {
		t.Fatal(err)
	}
	if ft.last != sel {
		t.Fatal("selections not forwarded")
	}
}

package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
)

type fakeTracker struct {
	last app.TrackOptions
	err  error
}

func (f *fakeTracker) RunTrackAuto(opts app.TrackOptions) error {
	f.last = opts
	return f.err
}

func (f *fakeTracker) RunTrackManual(_ app.TrackOptions) error {
	return errors.New("unexpected RunTrackManual")
}

func TestCompareIntent_Validate(t *testing.T) {
	t.Parallel()
	valid := app.TrackOptions{
		Baseline: "a", Current: "b", BenchmarkName: "Bench", ProfileType: "cpu",
		OutputFormat: "detailed", UseThreshold: false,
	}
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		i := &CompareIntent{Options: valid}
		if err := i.Validate(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("manual not allowed", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.IsManual = true
		i := &CompareIntent{Options: s}
		if err := i.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.OutputFormat = "nope"
		i := &CompareIntent{Options: s}
		if err := i.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("threshold without value", func(t *testing.T) {
		t.Parallel()
		s := valid
		s.UseThreshold = true
		s.RegressionThreshold = 0
		i := &CompareIntent{Options: s}
		if err := i.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCompareIntent_Run(t *testing.T) {
	t.Parallel()
	ft := &fakeTracker{}
	svc := &app.Services{Tracker: ft}
	opts := app.TrackOptions{
		Baseline: "base", Current: "cur", BenchmarkName: "B", ProfileType: "cpu",
		OutputFormat: "summary",
	}
	i := &CompareIntent{Options: opts}
	if err := RunValidated(i, svc); err != nil {
		t.Fatal(err)
	}
	if ft.last.Baseline != opts.Baseline || ft.last.Current != opts.Current {
		t.Fatal("options not forwarded")
	}
}

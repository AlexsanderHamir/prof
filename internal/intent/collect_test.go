package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
)

type fakeCollect struct {
	lastAuto   app.CollectAutoOptions
	lastManual app.CollectManualOptions
	err        error
}

func (f *fakeCollect) RunAuto(opts app.CollectAutoOptions) error {
	f.lastAuto = opts
	return f.err
}

func (f *fakeCollect) RunManual(opts app.CollectManualOptions) error {
	f.lastManual = opts
	return f.err
}

func (f *fakeCollect) DiscoverBenchmarks(string) ([]string, error) { return nil, nil }
func (f *fakeCollect) SupportedProfiles() []string                 { return nil }

func TestCollectIntent_Validate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		intent  CollectIntent
		wantErr bool
	}{
		{"ok", CollectIntent{Benchmarks: []string{"B"}, Profiles: []string{"cpu"}, Tag: "t", Count: 1}, false},
		{"no benches", CollectIntent{Profiles: []string{"cpu"}, Tag: "t", Count: 1}, true},
		{"no profiles", CollectIntent{Benchmarks: []string{"B"}, Tag: "t", Count: 1}, true},
		{"no tag", CollectIntent{Benchmarks: []string{"B"}, Profiles: []string{"cpu"}, Count: 1}, true},
		{"bad count", CollectIntent{Benchmarks: []string{"B"}, Profiles: []string{"cpu"}, Tag: "t", Count: 0}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.intent.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCollectIntent_Normalize(t *testing.T) {
	t.Parallel()
	i := &CollectIntent{
		Benchmarks: []string{"  B1 ", ""},
		Profiles:   []string{" cpu ", ""},
		Tag:        "  mytag  ",
		Count:      2,
	}
	i.Normalize()
	if got, want := i.Tag, "mytag"; got != want {
		t.Fatalf("Tag: got %q want %q", got, want)
	}
	if len(i.Benchmarks) != 1 || i.Benchmarks[0] != "B1" {
		t.Fatalf("Benchmarks: %#v", i.Benchmarks)
	}
	if len(i.Profiles) != 1 || i.Profiles[0] != "cpu" {
		t.Fatalf("Profiles: %#v", i.Profiles)
	}
}

func TestCollectIntent_Run(t *testing.T) {
	t.Parallel()
	fc := &fakeCollect{}
	svc := &app.Services{Collect: fc}
	intent := &CollectIntent{
		Benchmarks:      []string{"BenchA"},
		Profiles:        []string{"cpu", "memory"},
		Tag:             "v1",
		Count:           3,
		LenientProfiles: true,
		SkipPNG:         false,
	}
	if err := RunValidated(intent, svc); err != nil {
		t.Fatal(err)
	}
	if fc.lastAuto.Tag != "v1" || fc.lastAuto.Count != 3 {
		t.Fatalf("unexpected forwarding: tag=%q count=%d", fc.lastAuto.Tag, fc.lastAuto.Count)
	}
	if len(fc.lastAuto.Benchmarks) != 1 || fc.lastAuto.Benchmarks[0] != "BenchA" {
		t.Fatalf("benchmarks: %#v", fc.lastAuto.Benchmarks)
	}
	if len(fc.lastAuto.Profiles) != 2 {
		t.Fatalf("profiles: %#v", fc.lastAuto.Profiles)
	}
	if !fc.lastAuto.LenientProfiles || fc.lastAuto.SkipPNG {
		t.Fatalf("flags: lenient=%v skip=%v", fc.lastAuto.LenientProfiles, fc.lastAuto.SkipPNG)
	}
}

func TestCollectIntent_Run_propagatesError(t *testing.T) {
	t.Parallel()
	want := errors.New("boom")
	fc := &fakeCollect{err: want}
	svc := &app.Services{Collect: fc}
	intent := &CollectIntent{
		Benchmarks: []string{"B"},
		Profiles:   []string{"cpu"},
		Tag:        "t",
		Count:      1,
	}
	err := RunValidated(intent, svc)
	if !errors.Is(err, want) {
		t.Fatalf("err: %v", err)
	}
}

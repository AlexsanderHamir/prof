package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
)

type fakeBenchmark struct {
	lastBenchmarks      []string
	lastProfiles        []string
	lastTag             string
	lastCount           int
	lastGroupByPackage  bool
	lastLenientProfiles bool
	lastSkipPNG         bool
	err                 error
}

func (f *fakeBenchmark) RunBenchmarks(benchmarks, profiles []string, tag string, count int, groupByPackage, lenientProfiles, skipPNG bool) error {
	f.lastBenchmarks = append([]string(nil), benchmarks...)
	f.lastProfiles = append([]string(nil), profiles...)
	f.lastTag = tag
	f.lastCount = count
	f.lastGroupByPackage = groupByPackage
	f.lastLenientProfiles = lenientProfiles
	f.lastSkipPNG = skipPNG
	return f.err
}

func (f *fakeBenchmark) DiscoverBenchmarks(string) ([]string, error) { return nil, nil }

func (f *fakeBenchmark) SupportedProfiles() []string { return nil }

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
	fb := &fakeBenchmark{}
	svc := &app.Services{Benchmark: fb}
	intent := &CollectIntent{
		Benchmarks:      []string{"BenchA"},
		Profiles:        []string{"cpu", "memory"},
		Tag:             "v1",
		Count:           3,
		GroupByPackage:  true,
		LenientProfiles: true,
		SkipPNG:         false,
	}
	if err := RunValidated(intent, svc); err != nil {
		t.Fatal(err)
	}
	if fb.lastTag != "v1" || fb.lastCount != 3 {
		t.Fatalf("unexpected forwarding: tag=%q count=%d", fb.lastTag, fb.lastCount)
	}
	if len(fb.lastBenchmarks) != 1 || fb.lastBenchmarks[0] != "BenchA" {
		t.Fatalf("benchmarks: %#v", fb.lastBenchmarks)
	}
	if len(fb.lastProfiles) != 2 {
		t.Fatalf("profiles: %#v", fb.lastProfiles)
	}
	if !fb.lastGroupByPackage || !fb.lastLenientProfiles || fb.lastSkipPNG {
		t.Fatalf("flags: gbp=%v lenient=%v skip=%v", fb.lastGroupByPackage, fb.lastLenientProfiles, fb.lastSkipPNG)
	}
}

func TestCollectIntent_Run_propagatesError(t *testing.T) {
	t.Parallel()
	want := errors.New("boom")
	fb := &fakeBenchmark{err: want}
	svc := &app.Services{Benchmark: fb}
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

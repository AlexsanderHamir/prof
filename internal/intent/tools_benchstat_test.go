package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
)

type fakeTools struct {
	base, cur, bench string
	err              error
}

func (f *fakeTools) RunBenchStats(baseTag, currentTag, benchName string) error {
	f.base, f.cur, f.bench = baseTag, currentTag, benchName
	return f.err
}

func (f *fakeTools) RunQcacheGrind(_, _, _ string) error {
	return errors.New("unexpected qcachegrind")
}

func TestToolsBenchstatIntent_Validate(t *testing.T) {
	t.Parallel()
	if err := (&ToolsBenchstatIntent{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (&ToolsBenchstatIntent{BaseTag: "a", CurrentTag: "a", BenchName: "B"}).Validate(); err == nil {
		t.Fatal("expected error for same tags")
	}
}

func TestToolsBenchstatIntent_Run(t *testing.T) {
	t.Parallel()
	ft := &fakeTools{}
	svc := &app.Services{Tools: ft}
	i := &ToolsBenchstatIntent{BaseTag: "a", CurrentTag: "b", BenchName: "BenchX"}
	if err := RunValidated(i, svc); err != nil {
		t.Fatal(err)
	}
	if ft.base != "a" || ft.cur != "b" || ft.bench != "BenchX" {
		t.Fatalf("args: %q %q %q", ft.base, ft.cur, ft.bench)
	}
}

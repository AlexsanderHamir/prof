package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
)

type fakeToolsQ struct {
	tag, bench, prof string
	err              error
}

func (f *fakeToolsQ) RunBenchStats(_, _, _ string) error {
	return errors.New("unexpected benchstat")
}

func (f *fakeToolsQ) RunQcacheGrind(tag, benchName, profile string) error {
	f.tag, f.bench, f.prof = tag, benchName, profile
	return f.err
}

func TestToolsQcachegrindIntent_Validate(t *testing.T) {
	t.Parallel()
	if err := (&ToolsQcachegrindIntent{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestToolsQcachegrindIntent_Run(t *testing.T) {
	t.Parallel()
	ft := &fakeToolsQ{}
	svc := &app.Services{Tools: ft}
	i := &ToolsQcachegrindIntent{Tag: "t1", BenchName: "B", ProfileType: "cpu"}
	if err := RunValidated(i, svc); err != nil {
		t.Fatal(err)
	}
	if ft.tag != "t1" || ft.bench != "B" || ft.prof != "cpu" {
		t.Fatalf("args: %q %q %q", ft.tag, ft.bench, ft.prof)
	}
}

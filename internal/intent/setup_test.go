package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
)

type fakeSetup struct {
	calls int
	err   error
}

func (f *fakeSetup) CreateTemplate() error {
	f.calls++
	return f.err
}

func TestSetupIntent_Run(t *testing.T) {
	t.Parallel()
	fs := &fakeSetup{}
	svc := &app.Services{Setup: fs}
	i := &SetupIntent{}
	if err := RunValidated(i, svc); err != nil {
		t.Fatal(err)
	}
	if fs.calls != 1 {
		t.Fatalf("calls: %d", fs.calls)
	}
}

func TestSetupIntent_Run_error(t *testing.T) {
	t.Parallel()
	want := errors.New("write fail")
	fs := &fakeSetup{err: want}
	svc := &app.Services{Setup: fs}
	i := &SetupIntent{}
	err := RunValidated(i, svc)
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

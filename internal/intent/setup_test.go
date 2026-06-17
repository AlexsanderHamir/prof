package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
)

type fakeSetupConfig struct {
	calls int
	err   error
}

func (f *fakeSetupConfig) Load() (*config.Config, error) { return config.Default(""), nil }
func (f *fakeSetupConfig) Save(*config.Config) error     { return nil }
func (f *fakeSetupConfig) CreateDefaultFile() error {
	f.calls++
	return f.err
}
func (f *fakeSetupConfig) Path() (string, error) { return "", nil }

func TestSetupIntent_Run(t *testing.T) {
	t.Parallel()
	fc := &fakeSetupConfig{}
	svc := &app.Services{Config: fc}
	i := &SetupIntent{}
	if err := RunValidated(i, svc); err != nil {
		t.Fatal(err)
	}
	if fc.calls != 1 {
		t.Fatalf("calls: %d", fc.calls)
	}
}

func TestSetupIntent_Run_error(t *testing.T) {
	t.Parallel()
	want := errors.New("write fail")
	fc := &fakeSetupConfig{err: want}
	svc := &app.Services{Config: fc}
	i := &SetupIntent{}
	err := RunValidated(i, svc)
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

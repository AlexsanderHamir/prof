package intent

import (
	"errors"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
)

type fakeConfig struct {
	loadCalls   int
	saveCalls   int
	createCalls int
	saveErr     error
}

func (f *fakeConfig) Load() (*config.Config, error) {
	f.loadCalls++
	return config.Default(), nil
}

func (f *fakeConfig) Save(_ *config.Config) error {
	f.saveCalls++
	return f.saveErr
}

func (f *fakeConfig) CreateDefaultFile() error {
	f.createCalls++
	return nil
}

func (f *fakeConfig) Path() (string, error) {
	return "/tmp/prof.json", nil
}

func TestConfigCreateIntent_Run(t *testing.T) {
	t.Parallel()
	fc := &fakeConfig{}
	svc := &app.Services{Config: fc}
	if err := RunValidated(&ConfigCreateIntent{}, svc); err != nil {
		t.Fatal(err)
	}
	if fc.createCalls != 1 {
		t.Fatalf("create calls: %d", fc.createCalls)
	}
}

func TestConfigSaveIntent_Validate_nil(t *testing.T) {
	t.Parallel()
	err := (&ConfigSaveIntent{}).Validate()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigSaveIntent_Run(t *testing.T) {
	t.Parallel()
	fc := &fakeConfig{}
	svc := &app.Services{Config: fc}
	cfg := config.Default()
	if err := RunValidated(&ConfigSaveIntent{Config: cfg}, svc); err != nil {
		t.Fatal(err)
	}
	if fc.saveCalls != 1 {
		t.Fatalf("save calls: %d", fc.saveCalls)
	}
}

func TestConfigSaveIntent_Run_error(t *testing.T) {
	t.Parallel()
	want := errors.New("disk full")
	fc := &fakeConfig{saveErr: want}
	svc := &app.Services{Config: fc}
	err := RunValidated(&ConfigSaveIntent{Config: config.Default()}, svc)
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

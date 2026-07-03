package intent

import (
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

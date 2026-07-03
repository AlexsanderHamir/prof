package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
)

const (
	testProfCPU    = "cpu"
	testProfMemory = "memory"
)

type noopCollect struct{}

func (noopCollect) RunAuto(_ app.CollectAutoOptions) error        { return nil }
func (noopCollect) RunManual(_ app.CollectManualOptions) error    { return nil }
func (noopCollect) DiscoverBenchmarks(_ string) ([]string, error) { return nil, nil }
func (noopCollect) SupportedProfiles() []string                   { return nil }

type noopSetup struct{}

func (noopSetup) CreateTemplate() error { return nil }

type captureConfig struct{ createCalls int }

func (c *captureConfig) Load() (*config.Config, error) { return config.Default(), nil }
func (c *captureConfig) Save(*config.Config) error     { return nil }
func (c *captureConfig) CreateDefaultFile() error {
	c.createCalls++
	return nil
}
func (c *captureConfig) Path() (string, error) { return config.Path(config.Filename) }

type captureCollect struct {
	manual app.CollectManualOptions
	auto   app.CollectAutoOptions
}

func (c *captureCollect) RunAuto(opts app.CollectAutoOptions) error {
	c.auto = opts
	return nil
}

func (c *captureCollect) RunManual(opts app.CollectManualOptions) error {
	c.manual = opts
	return nil
}

func (*captureCollect) DiscoverBenchmarks(_ string) ([]string, error) { return nil, nil }
func (*captureCollect) SupportedProfiles() []string                   { return nil }

type errDiscoverCollect struct{ noopCollect }

func (errDiscoverCollect) DiscoverBenchmarks(string) ([]string, error) {
	return nil, errors.New("discover failed")
}

type emptyDiscoverCollect struct{ noopCollect }

func (emptyDiscoverCollect) DiscoverBenchmarks(string) ([]string, error) { return nil, nil }

func TestExecuteWithNilUsesOSArgs(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module cliexec\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"prof", "setup"}
	if err := ExecuteWith(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "prof.json")); err != nil {
		t.Fatal(err)
	}
}

func TestConfigInitCreatesProfJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module cliexec3\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	rootCmd := CreateRootCmd(nil)
	rootCmd.SetArgs([]string{"config", "init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "prof.json")); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteDelegatesToExecuteWithNil(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module cliexec2\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"prof", "setup"}
	if err := Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteWithStubServices(t *testing.T) {
	st := &captureConfig{}
	os.Args = []string{"prof", "config", "init"}
	err := ExecuteWith(&app.Services{
		Collect: noopCollect{},
		Config:  st,
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.createCalls != 1 {
		t.Fatalf("CreateDefaultFile calls=%d", st.createCalls)
	}
}

func TestCmdSetupRunE(t *testing.T) {
	st := &captureConfig{}
	root := CreateRootCmd(&app.Services{
		Collect: noopCollect{}, Config: st,
	})
	root.SetArgs([]string{"setup"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if st.createCalls != 1 {
		t.Fatal(st.createCalls)
	}
}

func TestCmdManualCollectRunE(t *testing.T) {
	captured := &captureCollect{}
	root := CreateRootCmd(&app.Services{
		Collect: captured, Setup: noopSetup{},
	})
	root.SetArgs([]string{CmdManual, "--tag", "t1", "--group-by-package", "a.prof", "b.prof"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if captured.manual.Tag != "t1" || !captured.manual.GroupByPackage || len(captured.manual.Files) != 2 || captured.manual.Files[0] != "a.prof" {
		t.Fatalf("%+v", captured.manual)
	}
}

func TestCmdAutoBenchmarkRunE(t *testing.T) {
	captured := &captureCollect{}
	root := CreateRootCmd(&app.Services{
		Collect: captured, Setup: noopSetup{},
	})
	root.SetArgs([]string{
		CmdAuto,
		"--benchmarks", "B1",
		"--profiles", testProfCPU + "," + testProfMemory,
		"--tag", "tg",
		"--count", "2",
		"--group-by-package",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if captured.auto.Tag != "tg" || captured.auto.Count != 2 || !captured.auto.GroupByPackage || len(captured.auto.Benchmarks) != 1 || captured.auto.Benchmarks[0] != "B1" {
		t.Fatalf("%+v", captured.auto)
	}
	if len(captured.auto.Profiles) != 2 || captured.auto.Profiles[0] != testProfCPU || captured.auto.Profiles[1] != testProfMemory {
		t.Fatalf("%+v", captured.auto)
	}
}

func TestCmdTuiRunEDiscoverError(t *testing.T) {
	rootMod := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootMod, "go.mod"), []byte("module tuierr\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(rootMod)
	root := CreateRootCmd(&app.Services{
		Collect: errDiscoverCollect{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{"tui"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "discover") {
		t.Fatalf("got %v", err)
	}
}

func TestCmdTuiRunENoBenchmarks(t *testing.T) {
	rootMod := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootMod, "go.mod"), []byte("module tuiempty\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(rootMod)
	root := CreateRootCmd(&app.Services{
		Collect: emptyDiscoverCollect{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{"tui"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "no benchmarks found") {
		t.Fatalf("got %v", err)
	}
}

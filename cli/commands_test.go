package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

const (
	testProfCPU    = "cpu"
	testProfMemory = "memory"
)

func resetCLIToolsGlobals(t *testing.T) {
	t.Helper()
	toolsGlobal = toolsFlags{}
}

type noopCollect struct{}

func (noopCollect) RunAuto(_ app.CollectAutoOptions) error        { return nil }
func (noopCollect) RunManual(_ app.CollectManualOptions) error    { return nil }
func (noopCollect) DiscoverBenchmarks(_ string) ([]string, error) { return nil, nil }
func (noopCollect) SupportedProfiles() []string                   { return nil }

type noopTrack struct{}

func (noopTrack) RunTrackAuto(_ app.TrackOptions) error   { return nil }
func (noopTrack) RunTrackManual(_ app.TrackOptions) error { return nil }

type noopTools struct{}

func (noopTools) RunBenchStats(_, _, _ string) error  { return nil }
func (noopTools) RunQcacheGrind(_, _, _ string) error { return nil }

type noopSetup struct{}

func (noopSetup) CreateTemplate() error { return nil }

func allNoopServices() *app.Services {
	return &app.Services{
		Collect: noopCollect{},
		Tracker: noopTrack{},
		Tools:   noopTools{},
		Setup:   noopSetup{},
	}
}

type captureSetup struct{ calls int }

func (c *captureSetup) CreateTemplate() error {
	c.calls++
	return nil
}

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

type captureTrack struct {
	auto, manual app.TrackOptions
}

func (c *captureTrack) RunTrackAuto(opts app.TrackOptions) error {
	c.auto = opts
	return nil
}

func (c *captureTrack) RunTrackManual(opts app.TrackOptions) error {
	c.manual = opts
	return nil
}

type captureTools struct {
	base, cur, bench string
	qTag, qBench     string
	qProf            string
}

func (c *captureTools) RunBenchStats(baseTag, currentTag, bench string) error {
	c.base, c.cur, c.bench = baseTag, currentTag, bench
	return nil
}

func (c *captureTools) RunQcacheGrind(tag, bench, profile string) error {
	c.qTag, c.qBench, c.qProf = tag, bench, profile
	return nil
}

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
	if _, err := os.Stat(filepath.Join(root, "config_template.json")); err != nil {
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
	st := &captureSetup{}
	os.Args = []string{"prof", "setup"}
	err := ExecuteWith(&app.Services{
		Collect: noopCollect{},
		Tracker: noopTrack{},
		Tools:   noopTools{},
		Setup:   st,
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.calls != 1 {
		t.Fatalf("CreateTemplate calls=%d", st.calls)
	}
}

func TestCmdSetupRunE(t *testing.T) {
	st := &captureSetup{}
	root := CreateRootCmd(&app.Services{
		Collect: noopCollect{}, Tracker: noopTrack{}, Tools: noopTools{}, Setup: st,
	})
	root.SetArgs([]string{"setup"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if st.calls != 1 {
		t.Fatal(st.calls)
	}
}

func TestCmdManualCollectRunE(t *testing.T) {
	captured := &captureCollect{}
	root := CreateRootCmd(&app.Services{
		Collect: captured, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
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
		Collect: captured, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
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

func TestCmdTrackAutoRunE(t *testing.T) {
	capturedTrack := &captureTrack{}
	root := CreateRootCmd(&app.Services{
		Collect: noopCollect{}, Tracker: capturedTrack, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{
		"track", CmdAuto,
		"--" + baseTagFlag, "base1",
		"--" + currentTagFlag, "cur1",
		"--" + benchNameFlag, "BenchX",
		"--profile-type", testProfCPU,
		"--output-format", "summary",
		"--fail-on-regression",
		"--regression-threshold", "5",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	s := capturedTrack.auto
	if s.Baseline != "base1" || s.Current != "cur1" || s.BenchmarkName != "BenchX" || s.ProfileType != testProfCPU ||
		s.OutputFormat != "summary" || !s.UseThreshold || s.RegressionThreshold != 5 || s.IsManual {
		t.Fatalf("%+v", s)
	}
}

func TestCmdTrackManualRunE(t *testing.T) {
	capturedTrack := &captureTrack{}
	root := CreateRootCmd(&app.Services{
		Collect: noopCollect{}, Tracker: capturedTrack, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{
		"track", CmdManual,
		"--" + baseTagFlag, "/b.txt",
		"--" + currentTagFlag, "/c.txt",
		"--output-format", "detailed-json",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	s := capturedTrack.manual
	if !s.IsManual || s.OutputFormat != "detailed-json" {
		t.Fatalf("%+v", s)
	}
}

func TestCmdToolsBenchstatAndQcachegrindRunE(t *testing.T) {
	capturedTools := &captureTools{}
	root := CreateRootCmd(&app.Services{
		Collect: noopCollect{}, Tracker: noopTrack{}, Tools: capturedTools, Setup: noopSetup{},
	})
	root.SetArgs([]string{"tools", workspace.ToolNameBenchstat, "--" + baseTagFlag, "a", "--" + currentTagFlag, "b", "--" + benchNameFlag, "B"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if capturedTools.base != "a" || capturedTools.cur != "b" || capturedTools.bench != "B" {
		t.Fatalf("%+v", capturedTools)
	}

	capturedTools2 := &captureTools{}
	root2 := CreateRootCmd(&app.Services{
		Collect: noopCollect{}, Tracker: noopTrack{}, Tools: capturedTools2, Setup: noopSetup{},
	})
	root2.SetArgs([]string{"tools", workspace.ToolNameQcachegrind, "--" + tagFlag, "t9", "--" + benchNameFlag, "BB", "--profiles", "mutex"})
	if err := root2.Execute(); err != nil {
		t.Fatal(err)
	}
	if capturedTools2.qTag != "t9" || capturedTools2.qBench != "BB" || capturedTools2.qProf != "mutex" {
		t.Fatalf("%+v", capturedTools2)
	}
}

func TestCmdTuiRunEDiscoverError(t *testing.T) {
	rootMod := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootMod, "go.mod"), []byte("module tuierr\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(rootMod)
	root := CreateRootCmd(&app.Services{
		Collect: errDiscoverCollect{}, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
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
		Collect: emptyDiscoverCollect{}, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{"tui"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "no benchmarks found") {
		t.Fatalf("got %v", err)
	}
}

func TestCmdTuiTrackRunENeedsTwoTags(t *testing.T) {
	rootMod := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootMod, "go.mod"), []byte("module tuitrack\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(rootMod)
	root := CreateRootCmd(allNoopServices())
	root.SetArgs([]string{"tui", "track"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "at least 2 tags") {
		t.Fatalf("got %v", err)
	}
}

func TestSetGlobalTrackingVariables(t *testing.T) {
	resetCLIToolsGlobals(t)
	setGlobalTrackingVariables(&app.TrackOptions{
		Baseline: "b", Current: "c", BenchmarkName: "bn", ProfileType: testProfCPU,
	})
	if toolsGlobal.baseline != "b" || toolsGlobal.current != "c" || toolsGlobal.benchmarkName != "bn" || toolsGlobal.profileType != testProfCPU {
		t.Fatal()
	}
}

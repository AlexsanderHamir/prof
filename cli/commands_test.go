package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/internal/app"
)

const (
	testProfCPU    = "cpu"
	testProfMemory = "memory"
)

func resetCLIPackageGlobals(t *testing.T) {
	t.Helper()
	benchmarks = nil
	profiles = nil
	tag = ""
	count = 0
	Baseline = ""
	Current = ""
	benchmarkName = ""
	profileType = ""
	outputFormat = ""
	failOnRegression = false
	regressionThreshold = 0
	groupByPackage = false
	lenientProfiles = false
	skipPNG = false
}

type noopBench struct{}

func (noopBench) RunBenchmarks(_, _ []string, _ string, _ int, _, _, _ bool) error { return nil }
func (noopBench) DiscoverBenchmarks(_ string) ([]string, error)                    { return nil, nil }
func (noopBench) SupportedProfiles() []string                                      { return nil }

type noopColl struct{}

func (noopColl) RunCollector(_ []string, _ string, _ bool) error { return nil }

type noopTrack struct{}

func (noopTrack) RunTrackAuto(_ *tracker.Selections) error   { return nil }
func (noopTrack) RunTrackManual(_ *tracker.Selections) error { return nil }

type noopTools struct{}

func (noopTools) RunBenchStats(_, _, _ string) error  { return nil }
func (noopTools) RunQcacheGrind(_, _, _ string) error { return nil }

type noopSetup struct{}

func (noopSetup) CreateTemplate() error { return nil }

func allNoopServices() *app.Services {
	return &app.Services{
		Benchmark: noopBench{},
		Collector: noopColl{},
		Tracker:   noopTrack{},
		Tools:     noopTools{},
		Setup:     noopSetup{},
	}
}

type captureSetup struct{ calls int }

func (c *captureSetup) CreateTemplate() error {
	c.calls++
	return nil
}

type captureColl struct {
	files []string
	tag   string
	group bool
}

func (c *captureColl) RunCollector(files []string, tag string, groupByPackage bool) error {
	c.files = append([]string(nil), files...)
	c.tag = tag
	c.group = groupByPackage
	return nil
}

type captureBench struct {
	bench, prof []string
	tag         string
	count       int
	group       bool
}

func (c *captureBench) RunBenchmarks(bench, prof []string, tag string, count int, groupByPackage bool, _, _ bool) error {
	c.bench = append([]string(nil), bench...)
	c.prof = append([]string(nil), prof...)
	c.tag = tag
	c.count = count
	c.group = groupByPackage
	return nil
}

func (*captureBench) DiscoverBenchmarks(_ string) ([]string, error) { return nil, nil }
func (*captureBench) SupportedProfiles() []string                   { return nil }

type errDiscoverBench struct{ noopBench }

func (errDiscoverBench) DiscoverBenchmarks(string) ([]string, error) {
	return nil, errors.New("discover failed")
}

type emptyDiscoverBench struct{ noopBench }

func (emptyDiscoverBench) DiscoverBenchmarks(string) ([]string, error) { return nil, nil }

type captureTrack struct {
	auto, manual *tracker.Selections
}

func (c *captureTrack) RunTrackAuto(sel *tracker.Selections) error {
	c.auto = sel
	return nil
}

func (c *captureTrack) RunTrackManual(sel *tracker.Selections) error {
	c.manual = sel
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
	resetCLIPackageGlobals(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module cliexec\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	old := os.Args
	t.Cleanup(func() {
		os.Args = old
		resetCLIPackageGlobals(t)
	})
	os.Args = []string{"prof", "setup"}
	if err := ExecuteWith(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "config_template.json")); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteDelegatesToExecuteWithNil(t *testing.T) {
	resetCLIPackageGlobals(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module cliexec2\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	old := os.Args
	t.Cleanup(func() {
		os.Args = old
		resetCLIPackageGlobals(t)
	})
	os.Args = []string{"prof", "setup"}
	if err := Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteWithStubServices(t *testing.T) {
	resetCLIPackageGlobals(t)
	st := &captureSetup{}
	old := os.Args
	t.Cleanup(func() {
		os.Args = old
		resetCLIPackageGlobals(t)
	})
	os.Args = []string{"prof", "setup"}
	err := ExecuteWith(&app.Services{
		Benchmark: noopBench{},
		Collector: noopColl{},
		Tracker:   noopTrack{},
		Tools:     noopTools{},
		Setup:     st,
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.calls != 1 {
		t.Fatalf("CreateTemplate calls=%d", st.calls)
	}
}

func TestCmdSetupRunE(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	st := &captureSetup{}
	root := CreateRootCmd(&app.Services{
		Benchmark: noopBench{}, Collector: noopColl{}, Tracker: noopTrack{}, Tools: noopTools{}, Setup: st,
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
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	capturedColl := &captureColl{}
	root := CreateRootCmd(&app.Services{
		Benchmark: noopBench{}, Collector: capturedColl, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{internal.MANUALCMD, "--tag", "t1", "--group-by-package", "a.prof", "b.prof"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if capturedColl.tag != "t1" || !capturedColl.group || len(capturedColl.files) != 2 || capturedColl.files[0] != "a.prof" {
		t.Fatalf("%+v", capturedColl)
	}
}

func TestCmdAutoBenchmarkRunE(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	capturedBench := &captureBench{}
	root := CreateRootCmd(&app.Services{
		Benchmark: capturedBench, Collector: noopColl{}, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{
		internal.AUTOCMD,
		"--benchmarks", "B1",
		"--profiles", testProfCPU + "," + testProfMemory,
		"--tag", "tg",
		"--count", "2",
		"--group-by-package",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if capturedBench.tag != "tg" || capturedBench.count != 2 || !capturedBench.group || len(capturedBench.bench) != 1 || capturedBench.bench[0] != "B1" {
		t.Fatalf("%+v", capturedBench)
	}
	if len(capturedBench.prof) != 2 || capturedBench.prof[0] != testProfCPU || capturedBench.prof[1] != testProfMemory {
		t.Fatalf("%+v", capturedBench)
	}
}

func TestCmdTrackAutoRunE(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	capturedTrack := &captureTrack{}
	root := CreateRootCmd(&app.Services{
		Benchmark: noopBench{}, Collector: noopColl{}, Tracker: capturedTrack, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{
		"track", internal.TrackAutoCMD,
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
	if s == nil {
		t.Fatal("nil selections")
	}
	if s.Baseline != "base1" || s.Current != "cur1" || s.BenchmarkName != "BenchX" || s.ProfileType != testProfCPU ||
		s.OutputFormat != "summary" || !s.UseThreshold || s.RegressionThreshold != 5 || s.IsManual {
		t.Fatalf("%+v", s)
	}
}

func TestCmdTrackManualRunE(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	capturedTrack := &captureTrack{}
	root := CreateRootCmd(&app.Services{
		Benchmark: noopBench{}, Collector: noopColl{}, Tracker: capturedTrack, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{
		"track", internal.TrackManualCMD,
		"--" + baseTagFlag, "/b.txt",
		"--" + currentTagFlag, "/c.txt",
		"--output-format", "detailed-json",
	})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	s := capturedTrack.manual
	if s == nil || !s.IsManual || s.OutputFormat != "detailed-json" {
		t.Fatalf("%+v", s)
	}
}

func TestCmdToolsBenchstatAndQcachegrindRunE(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	capturedTools := &captureTools{}
	root := CreateRootCmd(&app.Services{
		Benchmark: noopBench{}, Collector: noopColl{}, Tracker: noopTrack{}, Tools: capturedTools, Setup: noopSetup{},
	})
	root.SetArgs([]string{"tools", internal.ToolNameBenchstat, "--" + baseTagFlag, "a", "--" + currentTagFlag, "b", "--" + benchNameFlag, "B"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if capturedTools.base != "a" || capturedTools.cur != "b" || capturedTools.bench != "B" {
		t.Fatalf("%+v", capturedTools)
	}

	resetCLIPackageGlobals(t)
	capturedTools2 := &captureTools{}
	root2 := CreateRootCmd(&app.Services{
		Benchmark: noopBench{}, Collector: noopColl{}, Tracker: noopTrack{}, Tools: capturedTools2, Setup: noopSetup{},
	})
	root2.SetArgs([]string{"tools", internal.ToolNameQcachegrind, "--" + tagFlag, "t9", "--" + benchNameFlag, "BB", "--profiles", "mutex"})
	if err := root2.Execute(); err != nil {
		t.Fatal(err)
	}
	if capturedTools2.qTag != "t9" || capturedTools2.qBench != "BB" || capturedTools2.qProf != "mutex" {
		t.Fatalf("%+v", capturedTools2)
	}
}

func TestCmdTuiRunEDiscoverError(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	rootMod := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootMod, "go.mod"), []byte("module tuierr\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(rootMod)
	root := CreateRootCmd(&app.Services{
		Benchmark: errDiscoverBench{}, Collector: noopColl{}, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{"tui"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "discover") {
		t.Fatalf("got %v", err)
	}
}

func TestCmdTuiRunENoBenchmarks(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	rootMod := t.TempDir()
	if err := os.WriteFile(filepath.Join(rootMod, "go.mod"), []byte("module tuiempty\n\ngo 1.24.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(rootMod)
	root := CreateRootCmd(&app.Services{
		Benchmark: emptyDiscoverBench{}, Collector: noopColl{}, Tracker: noopTrack{}, Tools: noopTools{}, Setup: noopSetup{},
	})
	root.SetArgs([]string{"tui"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "no benchmarks found") {
		t.Fatalf("got %v", err)
	}
}

func TestCmdTuiTrackRunENeedsTwoTags(t *testing.T) {
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
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
	resetCLIPackageGlobals(t)
	t.Cleanup(func() { resetCLIPackageGlobals(t) })
	setGlobalTrackingVariables(&tracker.Selections{
		Baseline:            "b",
		Current:             "c",
		BenchmarkName:       "bn",
		ProfileType:         testProfCPU,
		OutputFormat:        "summary-json",
		UseThreshold:        true,
		RegressionThreshold: 4.5,
	})
	if Baseline != "b" || Current != "c" || benchmarkName != "bn" || profileType != testProfCPU ||
		outputFormat != "summary-json" || !failOnRegression || regressionThreshold != 4.5 {
		t.Fatal()
	}
}

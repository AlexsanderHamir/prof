package collect

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/testpaths"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func setupProcessProfilesEnv(t *testing.T, tag, bench string, profiles []string) (workspace.TagLayout, string) {
	t.Helper()
	fixture := testpaths.MustAsset(t, "fixtures", filterFixtureCPU)
	modRoot := t.TempDir()
	writeModuleRoot(t, modRoot)
	t.Chdir(modRoot)

	if err := setupDirectories(tag, []string{bench}, profiles, false); err != nil {
		t.Fatal(err)
	}
	layout, err := workspace.TagLayoutFromCWD(tag)
	if err != nil {
		t.Fatal(err)
	}
	return layout, fixture
}

func copyFixtureToProfile(t *testing.T, layout workspace.TagLayout, bench, profile, src string) {
	t.Helper()
	dst := layout.ProfileBinary(bench, profile)
	if err := os.MkdirAll(filepath.Dir(dst), workspace.PermDir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(dst, data, workspace.PermFile); err != nil {
		t.Fatal(err)
	}
}

func TestProcessProfiles_skipsMissingBinary(t *testing.T) {
	const (
		tag   = "t1"
		bench = "BenchmarkFoo"
	)
	layout, fixture := setupProcessProfilesEnv(t, tag, bench, []string{"cpu", "memory"})
	copyFixtureToProfile(t, layout, bench, "cpu", fixture)

	runner := &tooling.FakeRunner{
		Out: [][]byte{[]byte("flat profile text"), []byte("tree profile text"), []byte("png-bytes")},
	}
	processed, err := processProfiles(runner, bench, []string{"cpu", "memory"}, tag, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(processed) != 1 || processed[0] != "cpu" {
		t.Fatalf("processed: %#v", processed)
	}
}

func TestProcessProfiles_failsWhenAllMissing(t *testing.T) {
	const (
		tag   = "t2"
		bench = "BenchmarkFoo"
	)
	_, _ = setupProcessProfilesEnv(t, tag, bench, []string{"cpu", "memory"})

	_, err := processProfiles(&tooling.FakeRunner{}, bench, []string{"cpu", "memory"}, tag, nil)
	if err == nil {
		t.Fatal("expected error when no profile binaries exist")
	}
}

func TestProcessProfiles_continuesOnPNGFailure(t *testing.T) {
	const (
		tag   = "t3"
		bench = "BenchmarkFoo"
	)
	layout, fixture := setupProcessProfilesEnv(t, tag, bench, []string{"cpu"})
	copyFixtureToProfile(t, layout, bench, "cpu", fixture)

	runner := &tooling.FakeRunner{
		Out: [][]byte{[]byte("flat profile text"), []byte("tree profile text")},
		Err: []error{nil, nil, errors.New("graphviz unavailable")},
	}
	processed, err := processProfiles(runner, bench, []string{"cpu"}, tag, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(processed) != 1 || processed[0] != "cpu" {
		t.Fatalf("processed: %#v", processed)
	}
	if _, statErr := os.Stat(layout.Hotspot(bench, "cpu")); statErr != nil {
		t.Fatalf("expected hotspot summary: %v", statErr)
	}
	if _, statErr := os.Stat(layout.CallTreeText(bench, "cpu")); statErr != nil {
		t.Fatalf("expected call tree text: %v", statErr)
	}
	if _, statErr := os.Stat(layout.CallTreeJSON(bench, "cpu")); statErr != nil {
		t.Fatalf("expected call tree json: %v", statErr)
	}
}

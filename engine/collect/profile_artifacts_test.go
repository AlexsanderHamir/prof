package collect

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/testpaths"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestProfileArtifacts_catalogOrder(t *testing.T) {
	arts := profileArtifacts()
	if len(arts) != 4 {
		t.Fatalf("expected 4 artifacts, got %d", len(arts))
	}
	want := []string{artifactHotspots, artifactCallTreeText, artifactCallTreeJSON, artifactCallGraphPNG}
	for i, id := range want {
		if arts[i].ID != id {
			t.Fatalf("artifact[%d]=%q want %q", i, arts[i].ID, id)
		}
	}
	if arts[3].Policy != BestEffort {
		t.Fatalf("png policy=%v want BestEffort", arts[3].Policy)
	}
}

func TestEmitProfileArtifactsFromCatalog_requiredFailure(t *testing.T) {
	const (
		tag   = "catalog-req"
		bench = "BenchmarkFoo"
	)
	modRoot := t.TempDir()
	writeModuleRoot(t, modRoot)
	t.Chdir(modRoot)
	if err := setupDirectories(tag, []string{bench}, []string{"cpu"}, false); err != nil {
		t.Fatal(err)
	}
	layout, err := workspace.TagLayoutFromCWD(tag)
	if err != nil {
		t.Fatal(err)
	}
	runner := &tooling.FakeRunner{
		Err: []error{errors.New("pprof top failed")},
	}
	ctx := ProduceContext{
		Runner:  runner,
		Layout:  layout,
		Bench:   bench,
		Profile: "cpu",
		BinPath: "cpu.out",
		Bundle:  &parser.ProfileBundle{CallGraph: &parser.CallGraphData{Total: 1}},
	}
	if emitErr := emitProfileArtifactsFromCatalog(ctx); emitErr == nil {
		t.Fatal("expected required hotspot failure")
	}
}

func TestEmitProfileArtifactsFromCatalog_bestEffortPNG(t *testing.T) {
	const (
		tag   = "catalog-png"
		bench = "BenchmarkFoo"
	)
	fixture := testpaths.MustAsset(t, "fixtures", filterFixtureCPU)
	modRoot := t.TempDir()
	writeModuleRoot(t, modRoot)
	t.Chdir(modRoot)
	if err := setupDirectories(tag, []string{bench}, []string{"cpu"}, false); err != nil {
		t.Fatal(err)
	}
	layout, err := workspace.TagLayoutFromCWD(tag)
	if err != nil {
		t.Fatal(err)
	}
	dst := layout.ProfileBinary(bench, "cpu")
	if mkdirErr := os.MkdirAll(filepath.Dir(dst), workspace.PermDir); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(dst, data, workspace.PermFile); err != nil {
		t.Fatal(err)
	}

	bundle, err := parser.BundleFromPath(dst)
	if err != nil {
		t.Fatal(err)
	}
	runner := &tooling.FakeRunner{
		Out: [][]byte{[]byte("top"), []byte("tree")},
		Err: []error{nil, nil, errors.New("graphviz unavailable")},
	}
	ctx := ProduceContext{
		Runner:  runner,
		Layout:  layout,
		Bench:   bench,
		Profile: "cpu",
		BinPath: dst,
		Bundle:  bundle,
	}
	if emitErr := emitProfileArtifactsFromCatalog(ctx); emitErr != nil {
		t.Fatal(emitErr)
	}
	for _, path := range []string{
		layout.Hotspot(bench, "cpu"),
		layout.CallTreeText(bench, "cpu"),
		layout.CallTreeJSON(bench, "cpu"),
	} {
		if _, statErr := os.Stat(path); statErr != nil {
			t.Fatalf("missing %s: %v", path, statErr)
		}
	}
}

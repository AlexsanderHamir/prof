package datamap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestBuild_minimalCPUProfile(t *testing.T) {
	t.Parallel()
	modRoot := t.TempDir()
	layout := workspace.NewTagLayout(modRoot, "baseline")
	bench := "BenchmarkFoo"

	snap := ProfileSnapshot{
		Profile: "cpu",
		ProfileData: &parser.ProfileData{
			Total: 100,
			Flat:  map[string]int64{"main.work": 30},
			Cum:   map[string]int64{"main.work": 80},
			FlatPercentages: map[string]float64{
				"main.work": 30,
			},
			CumPercentages: map[string]float64{
				"main.work": 80,
			},
			SortedEntries: []parser.FuncEntry{
				{Name: "main.work", Flat: 30},
			},
		},
		ListEntries: []parser.FunctionListEntry{
			{OutputStem: "work", FullSymbol: "main.work"},
		},
		SourceLinesCollected: 1,
	}

	m, err := Build(BuildInput{
		Layout:         layout,
		Tag:            "baseline",
		Benchmark:      bench,
		CollectionMode: collectionManual,
		Profiles:       []string{"cpu"},
		PerProfile:     []ProfileSnapshot{snap},
	})
	if err != nil {
		t.Fatal(err)
	}

	if m.SchemaVersion != SchemaVersion {
		t.Fatalf("schema_version=%d", m.SchemaVersion)
	}
	if got := m.Hotspots["cpu"].Path; got != "hotspots/BenchmarkFoo/cpu.txt" {
		t.Fatalf("hotspot path=%q", got)
	}
	fn := m.SourceLines["cpu"].Functions["work"]
	if fn.FullSymbol != "main.work" || fn.Status != statusOK {
		t.Fatalf("function ref=%+v", fn)
	}
	if len(m.Hotspots["cpu"].TopSymbols) != 1 {
		t.Fatalf("top symbols=%d", len(m.Hotspots["cpu"].TopSymbols))
	}
}

func TestWriteJSON_roundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	layout := workspace.NewTagLayout(dir, "t1")
	path := layout.DataMapping("BenchmarkBar")

	m := BenchmarkMap{
		SchemaVersion: SchemaVersion,
		Tag:           "t1",
		Benchmark:     "BenchmarkBar",
		Profiles:      map[string]ProfileRef{},
		Hotspots:      map[string]HotspotSection{},
		CallTrees:     map[string]CallTreeSection{},
		SourceLines:   map[string]SourceLinesSection{},
		Status: Status{
			Profiles:    map[string]string{},
			Hotspots:    map[string]string{},
			CallTrees:   map[string]string{},
			SourceLines: map[string]SourceLinesStatus{},
		},
	}
	if err := WriteJSON(path, m); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"schema_version": 1`) {
		t.Fatalf("json=%s", data)
	}
}

func TestParseMeasurementSummary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "run.txt")
	content := `goos: linux
BenchmarkFoo-8    1000    12345 ns/op    100 B/op    2 allocs/op
PASS
ok  	example.com/bench	1.234s
`
	if err := os.WriteFile(path, []byte(content), workspace.PermFile); err != nil {
		t.Fatal(err)
	}
	sum, err := parseMeasurementSummary(path)
	if err != nil {
		t.Fatal(err)
	}
	if sum.NsPerOpMedian != 12345 || sum.BytesPerOp != 100 || sum.AllocsPerOp != 2 {
		t.Fatalf("summary=%+v", sum)
	}
	if sum.Result != benchResultPass {
		t.Fatalf("result=%q", sum.Result)
	}
}

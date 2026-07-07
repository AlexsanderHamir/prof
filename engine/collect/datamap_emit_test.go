package collect

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/datamap"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestEmitBenchmarkMap_writesMapJSON(t *testing.T) {
	const (
		tag   = "maptest"
		bench = "BenchmarkFoo"
	)
	layout, fixture := setupProcessProfilesEnv(t, tag, []string{"cpu"})
	copyFixtureToProfile(t, layout, bench, "cpu", fixture)

	runner := &tooling.FakeRunner{
		Out: [][]byte{
			[]byte("flat profile text"),
			[]byte("tree profile text"),
			[]byte("png-bytes"),
		},
	}
	processed, err := processProfiles(runner, bench, []string{"cpu"}, tag, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(processed) != 1 {
		t.Fatalf("processed=%v", processed)
	}

	cpuPath := layout.ProfileBinary(bench, "cpu")
	filter := config.FunctionFilter{}
	entries, profileData, err := parser.GetFunctionListEntriesWithProfileData(cpuPath, filter)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected function entries from fixture")
	}

	emitBenchmarkMap(nil, layout, emitMapParams{
		Tag:            tag,
		Benchmark:      bench,
		Profiles:       processed,
		Filter:         filter,
		CollectionMode: datamapCollectionAuto,
		PerProfile: []datamap.ProfileSnapshot{{
			Profile:              "cpu",
			ProfileData:          profileData,
			ListEntries:          entries,
			SourceLinesCollected: len(entries),
		}},
		IncludeMeasuring: false,
	})

	mapPath := layout.DataMapping(bench)
	data, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	var m datamap.BenchmarkMap
	if err = json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != datamap.SchemaVersion {
		t.Fatalf("schema_version=%d", m.SchemaVersion)
	}
	if m.Hotspots["cpu"].Path != "hotspots/BenchmarkFoo/cpu.txt" {
		t.Fatalf("hotspot path=%q", m.Hotspots["cpu"].Path)
	}
	if len(m.SourceLines["cpu"].Functions) == 0 {
		t.Fatal("expected source_lines functions")
	}
	if len(m.Hotspots["cpu"].TopSymbols) == 0 {
		t.Fatal("expected top_symbols")
	}
}

package parser

import (
	"encoding/json"
	"os"
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestBuildCallGraphFromProfile_syntheticStack(t *testing.T) {
	callerFn := &pprofprofile.Function{Name: "main.main"}
	calleeFn := &pprofprofile.Function{Name: "pkg.Work"}
	callerLoc := &pprofprofile.Location{Line: []pprofprofile.Line{{Function: callerFn}}}
	calleeLoc := &pprofprofile.Location{Line: []pprofprofile.Line{{Function: calleeFn}}}
	p := &pprofprofile.Profile{
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Sample: []*pprofprofile.Sample{
			{
				Location: []*pprofprofile.Location{calleeLoc, callerLoc},
				Value:    []int64{10},
			},
		},
	}
	cg := BuildCallGraphFromProfile(p, 0)
	if cg.Total != 10 {
		t.Fatalf("total=%d", cg.Total)
	}
	if len(cg.Nodes) != 2 {
		t.Fatalf("nodes=%d", len(cg.Nodes))
	}
	if len(cg.Edges) != 1 {
		t.Fatalf("edges=%#v", cg.Edges)
	}
	e := cg.Edges[0]
	if e.Caller != "main.main" || e.Callee != "pkg.Work" || e.Weight != 10 {
		t.Fatalf("edge=%#v", e)
	}
}

func TestCallGraphFromPath_fixture(t *testing.T) {
	path := benchmarkGenPoolCPUFixturePath(t)
	cg, err := CallGraphFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if cg.Total <= 0 {
		t.Fatal("expected positive total")
	}
	if len(cg.Nodes) == 0 {
		t.Fatal("expected nodes")
	}

	dir := t.TempDir()
	out := dir + "/cg.json"
	if writeErr := WriteCallGraphJSON(out, cg); writeErr != nil {
		t.Fatal(writeErr)
	}
	raw, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Fatal(readErr)
	}
	var decoded CallGraphData
	if unmarshalErr := json.Unmarshal(raw, &decoded); unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}
	if decoded.Total != cg.Total || len(decoded.Nodes) != len(cg.Nodes) {
		t.Fatalf("round-trip mismatch total=%d nodes=%d", decoded.Total, len(decoded.Nodes))
	}
}

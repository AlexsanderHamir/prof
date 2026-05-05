package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
	pprofprofile "github.com/google/pprof/profile"
)

// TestEdge_functionListCollection_fixture exercises collector.GetFunctionsOutput
// against the committed CPU fixture using a real profile symbol whose FullSymbol
// contains regexp metacharacters (parentheses from pointer receiver syntax). That
// path relies on regexp.QuoteMeta in the collector.
func TestEdge_functionListCollection_fixture(t *testing.T) {
	cpuPath := edgecasesFixturePath(t, fixtureCPUFile)
	entries, err := parser.GetFunctionListEntriesV2(cpuPath, internal.FunctionFilter{})
	if err != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", err)
	}

	var pick parser.FunctionListEntry
	for _, e := range entries {
		if strings.Contains(e.FullSymbol, "(") && strings.Contains(e.FullSymbol, ")") {
			pick = e
			break
		}
	}
	if pick.FullSymbol == "" {
		t.Fatal("fixture should include at least one symbol with '(' and ')' in FullSymbol")
	}

	dir := t.TempDir()
	if err := collector.GetFunctionsOutput([]parser.FunctionListEntry{pick}, cpuPath, dir); err != nil {
		t.Fatalf("GetFunctionsOutput: %v", err)
	}
	out := filepath.Join(dir, pick.OutputStem+"."+internal.TextExtension)
	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("expected per-function list file %s: %v", out, err)
	}
	if st.Size() == 0 {
		t.Fatalf("expected non-empty list output for %q", pick.FullSymbol)
	}
}

// TestEdge_functionListCollection_renamedFixtureSymbol clones the committed CPU
// fixture, rewrites one function's name to include regexp metacharacters, and
// asserts the parser pipeline still extracts the expected short stem. This
// complements TestEdge_functionListCollection_fixture without requiring a
// hand-built protobuf profile (which is easy to get subtly wrong).
func TestEdge_functionListCollection_renamedFixtureSymbol(t *testing.T) {
	const weirdName = `edge/syn.(*Re[go.shape.string]).Method`
	cpuPath := edgecasesFixturePath(t, fixtureCPUFile)
	raw, err := os.ReadFile(cpuPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	p, err := pprofprofile.Parse(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Parse fixture: %v", err)
	}

	renamed := false
outer:
	for _, s := range p.Sample {
		for _, loc := range s.Location {
			if loc == nil {
				continue
			}
			for i := range loc.Line {
				if loc.Line[i].Function != nil {
					loc.Line[i].Function.Name = weirdName
					renamed = true
					break outer
				}
			}
		}
	}
	if !renamed {
		t.Fatal("could not find a function line to rename")
	}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := pprofprofile.Parse(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("round-trip parse after rename: %v", err)
	}

	tmp := filepath.Join(t.TempDir(), "mutated_edge.out")
	if err := os.WriteFile(tmp, buf.Bytes(), internal.PermFile); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries, err := parser.GetFunctionListEntriesV2(tmp, internal.FunctionFilter{})
	if err != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.FullSymbol == weirdName {
			found = true
			if got, want := e.OutputStem, "Method"; got != want {
				t.Fatalf("OutputStem: got %q want %q", got, want)
			}
			break
		}
	}
	if !found {
		t.Fatalf("renamed symbol %q not present in entries", weirdName)
	}
}

package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/workspace"

	"github.com/AlexsanderHamir/prof/engine/collect"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/parser"
	pprofprofile "github.com/google/pprof/profile"
)

// edgecasesRenameFirstNamedFunction sets the first non-nil function name in p.
func edgecasesRenameFirstNamedFunction(p *pprofprofile.Profile, name string) bool {
	for _, s := range p.Sample {
		for _, loc := range s.Location {
			if loc == nil {
				continue
			}
			for i := range loc.Line {
				if loc.Line[i].Function != nil {
					loc.Line[i].Function.Name = name
					return true
				}
			}
		}
	}
	return false
}

// edgecasesWriteProfileRoundTrip writes p to a temp file after verifying parse round-trip.
func edgecasesWriteProfileRoundTrip(t *testing.T, p *pprofprofile.Profile, fileName string) string {
	t.Helper()
	var buf bytes.Buffer
	if werr := p.Write(&buf); werr != nil {
		t.Fatalf("Write: %v", werr)
	}
	if _, rerr := pprofprofile.Parse(bytes.NewReader(buf.Bytes())); rerr != nil {
		t.Fatalf("round-trip parse after rename: %v", rerr)
	}
	tmp := filepath.Join(t.TempDir(), fileName)
	if wferr := os.WriteFile(tmp, buf.Bytes(), workspace.PermFile); wferr != nil {
		t.Fatalf("WriteFile: %v", wferr)
	}
	return tmp
}

func edgecasesFindEntryByFullSymbol(entries []parser.FunctionListEntry, full string) (parser.FunctionListEntry, bool) {
	for _, e := range entries {
		if e.FullSymbol == full {
			return e, true
		}
	}
	return parser.FunctionListEntry{}, false
}

// TestEdge_functionListCollection_fixture exercises collector.GetFunctionsOutput
// against the committed CPU fixture using a real profile symbol whose FullSymbol
// contains regexp metacharacters (parentheses from pointer receiver syntax). That
// path relies on regexp.QuoteMeta in the collector.
func TestEdge_functionListCollection_fixture(t *testing.T) {
	cpuPath := edgecasesFixturePath(t, fixtureCPUFile)
	entries, listErr := parser.GetFunctionListEntriesV2(cpuPath, config.FunctionFilter{})
	if listErr != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", listErr)
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
	if outErr := collect.FunctionsOutput(tooling.NewExecRunner(), []parser.FunctionListEntry{pick}, cpuPath, dir); outErr != nil {
		t.Fatalf("GetFunctionsOutput: %v", outErr)
	}
	out := filepath.Join(dir, pick.OutputStem+"."+workspace.TextExtension)
	st, statErr := os.Stat(out)
	if statErr != nil {
		t.Fatalf("expected per-function list file %s: %v", out, statErr)
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
	raw, readErr := os.ReadFile(cpuPath)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	p, parseErr := pprofprofile.Parse(bytes.NewReader(raw))
	if parseErr != nil {
		t.Fatalf("Parse fixture: %v", parseErr)
	}
	if !edgecasesRenameFirstNamedFunction(p, weirdName) {
		t.Fatal("could not find a function line to rename")
	}

	tmp := edgecasesWriteProfileRoundTrip(t, p, "mutated_edge.out")
	entries, listErr := parser.GetFunctionListEntriesV2(tmp, config.FunctionFilter{})
	if listErr != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", listErr)
	}
	e, ok := edgecasesFindEntryByFullSymbol(entries, weirdName)
	if !ok {
		t.Fatalf("renamed symbol %q not present in entries", weirdName)
	}
	if got, want := e.OutputStem, "Method"; got != want {
		t.Fatalf("OutputStem: got %q want %q", got, want)
	}
}

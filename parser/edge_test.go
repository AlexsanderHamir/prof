package parser_test

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/testpaths"

	"github.com/AlexsanderHamir/prof/parser"
)

const (
	edgeBenchName   = "BenchmarkStringProcessor"
	edgeFuncProcess = "ProcessStrings"
	edgeFixtureCPU  = edgeBenchName + "_cpu.out"
	edgeFixtureMem  = edgeBenchName + "_memory.out"
)

func edgeFixturePath(t *testing.T, fileName string) string {
	t.Helper()
	return testpaths.MustAsset(t, "fixtures", fileName)
}

func TestEdge_profileReader_errors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		body    string
		wantSub string
	}{
		{
			name:    "empty_reader",
			body:    "",
			wantSub: "parse",
		},
		{
			name:    "non_profile_bytes",
			body:    "not-a-pprof-protobuf-payload",
			wantSub: "parse",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parser.ProfileDataFromReader(strings.NewReader(tc.body))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(strings.ToLower(err.Error()), tc.wantSub) {
				t.Fatalf("error %q should mention %q", err.Error(), tc.wantSub)
			}
		})
	}
}

func TestEdge_profileReader_missingPath(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "does-not-exist-profile.out")
	_, err := parser.GetFunctionListEntriesV2(missing, config.FunctionFilter{})
	if err == nil {
		t.Fatal("expected error for missing profile path")
	}
}

func TestEdge_functionListEntries_includeMatchesNothing(t *testing.T) {
	t.Parallel()
	cpuPath := edgeFixturePath(t, edgeFixtureCPU)
	f := config.FunctionFilter{
		IncludePrefixes: []string{"import/path/that/cannot/exist/in/fixture/zzzz"},
	}
	entries, err := parser.GetFunctionListEntriesV2(cpuPath, f)
	if err != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("want no entries, got %d (first %q)", len(entries), entries[0].OutputStem)
	}
}

func TestEdge_functionListEntries_duplicateIgnoreSameAsSingle(t *testing.T) {
	t.Parallel()
	cpuPath := edgeFixturePath(t, edgeFixtureCPU)
	once := config.FunctionFilter{IgnoreFunctions: []string{edgeBenchName}}
	dup := config.FunctionFilter{IgnoreFunctions: []string{edgeBenchName, edgeBenchName}}

	a, err := parser.GetFunctionListEntriesV2(cpuPath, once)
	if err != nil {
		t.Fatalf("GetFunctionListEntriesV2 once: %v", err)
	}
	b, err2 := parser.GetFunctionListEntriesV2(cpuPath, dup)
	if err2 != nil {
		t.Fatalf("GetFunctionListEntriesV2 dup: %v", err2)
	}
	if len(a) != len(b) {
		t.Fatalf("len mismatch: once=%d dup=%d", len(a), len(b))
	}
	sortEntriesByFullSymbol(a)
	sortEntriesByFullSymbol(b)
	for i := range a {
		if a[i].OutputStem != b[i].OutputStem || a[i].FullSymbol != b[i].FullSymbol {
			t.Fatalf("row %d differs: %#v vs %#v", i, a[i], b[i])
		}
	}
}

func TestEdge_functionListEntries_ignoreWithoutIncludePrefixes(t *testing.T) {
	t.Parallel()
	memPath := edgeFixturePath(t, edgeFixtureMem)
	none := config.FunctionFilter{}
	withIgnore := config.FunctionFilter{IgnoreFunctions: []string{edgeFuncProcess}}

	all, err := parser.GetFunctionListEntriesV2(memPath, none)
	if err != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", err)
	}
	filtered, err2 := parser.GetFunctionListEntriesV2(memPath, withIgnore)
	if err2 != nil {
		t.Fatalf("GetFunctionListEntriesV2: %v", err2)
	}
	if len(filtered) >= len(all) {
		t.Fatalf("expected fewer entries after ignore, all=%d filtered=%d", len(all), len(filtered))
	}
	for _, e := range filtered {
		if e.OutputStem == edgeFuncProcess {
			t.Fatalf("ignored short name still present: %q", e.OutputStem)
		}
	}
}

func sortEntriesByFullSymbol(entries []parser.FunctionListEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].FullSymbol != entries[j].FullSymbol {
			return entries[i].FullSymbol < entries[j].FullSymbol
		}
		return entries[i].OutputStem < entries[j].OutputStem
	})
}

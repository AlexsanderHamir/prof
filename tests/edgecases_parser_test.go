package tests

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func TestEdge_profileReader_errors(t *testing.T) {
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
	_, err := parser.GetFunctionListEntriesV2(missing, internal.FunctionFilter{})
	if err == nil {
		t.Fatal("expected error for missing profile path")
	}
}

func TestEdge_functionListEntries_filterMatrix(t *testing.T) {
	cpuPath := edgecasesFixturePath(t, fixtureCPUFile)

	t.Run("include_prefixes_match_nothing", func(t *testing.T) {
		t.Parallel()
		f := internal.FunctionFilter{
			IncludePrefixes: []string{"import/path/that/cannot/exist/in/fixture/zzzz"},
		}
		entries, err := parser.GetFunctionListEntriesV2(cpuPath, f)
		if err != nil {
			t.Fatalf("GetFunctionListEntriesV2: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("want no entries, got %d (first %q)", len(entries), entries[0].OutputStem)
		}
	})

	t.Run("duplicate_ignore_functions_same_as_single", func(t *testing.T) {
		t.Parallel()
		once := internal.FunctionFilter{IgnoreFunctions: []string{benchName}}
		dup := internal.FunctionFilter{IgnoreFunctions: []string{benchName, benchName}}

		a, err := parser.GetFunctionListEntriesV2(cpuPath, once)
		if err != nil {
			t.Fatalf("GetFunctionListEntriesV2 once: %v", err)
		}
		b, err := parser.GetFunctionListEntriesV2(cpuPath, dup)
		if err != nil {
			t.Fatalf("GetFunctionListEntriesV2 dup: %v", err)
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
	})

	t.Run("empty_include_prefixes_still_applies_ignore", func(t *testing.T) {
		t.Parallel()
		memPath := edgecasesFixturePath(t, fixtureMemFile)
		none := internal.FunctionFilter{}
		withIgnore := internal.FunctionFilter{IgnoreFunctions: []string{funcProcess}}

		all, err := parser.GetFunctionListEntriesV2(memPath, none)
		if err != nil {
			t.Fatalf("GetFunctionListEntriesV2: %v", err)
		}
		filtered, err := parser.GetFunctionListEntriesV2(memPath, withIgnore)
		if err != nil {
			t.Fatalf("GetFunctionListEntriesV2: %v", err)
		}
		if len(filtered) >= len(all) {
			t.Fatalf("expected fewer entries after ignore, all=%d filtered=%d", len(all), len(filtered))
		}
		for _, e := range filtered {
			if e.OutputStem == funcProcess {
				t.Fatalf("ignored short name still present: %q", e.OutputStem)
			}
		}
	})
}

func sortEntriesByFullSymbol(entries []parser.FunctionListEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].FullSymbol != entries[j].FullSymbol {
			return entries[i].FullSymbol < entries[j].FullSymbol
		}
		return entries[i].OutputStem < entries[j].OutputStem
	})
}

package collect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/testpaths"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

const filterFixtureCPU = "BenchmarkStringProcessor_cpu.out"

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestWriteArtifactFile_createsParent(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "nested", "artifact.txt")
	if err := writeArtifactFile(out, []byte("payload")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "payload" {
		t.Fatalf("got %q", data)
	}
}

func TestGetProfileTextOutput_fakeRunner(t *testing.T) {
	t.Parallel()
	runner := &tooling.FakeRunner{
		Out: [][]byte{[]byte("flat  flat%   sum%        cum   cum%")},
	}
	out := filepath.Join(t.TempDir(), "top.txt")
	if err := getProfileTextOutput(runner, "cpu.out", out); err != nil {
		t.Fatal(err)
	}
	if len(runner.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runner.Runs))
	}
	if !strings.Contains(string(mustReadFile(t, out)), "flat") {
		t.Fatal("expected pprof text in output file")
	}
}

func TestGetProfileTextOutput_nilRunner(t *testing.T) {
	t.Parallel()
	if err := getProfileTextOutput(nil, "cpu.out", "out.txt"); err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestWriteFunctionListPprof_fakeRunner(t *testing.T) {
	t.Parallel()
	runner := &tooling.FakeRunner{
		Out: [][]byte{[]byte("ROUTINE ======================== ProcessStrings")},
	}
	out := filepath.Join(t.TempDir(), "fn.txt")
	if err := writeFunctionListPprof(runner, "ProcessStrings", "pkg.ProcessStrings", "cpu.out", out); err != nil {
		t.Fatal(err)
	}
	if len(runner.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runner.Runs))
	}
}

func TestGetFunctionsOutput_fakeRunner(t *testing.T) {
	t.Parallel()
	cpuPath := testpaths.MustAsset(t, "fixtures", filterFixtureCPU)
	entries, err := parser.GetFunctionListEntriesV2(cpuPath, config.FunctionFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("fixture should yield function entries")
	}

	pick := entries[0]
	runner := &tooling.FakeRunner{
		Out: [][]byte{[]byte("list output for " + pick.OutputStem)},
	}
	dir := t.TempDir()
	if outErr := getFunctionsOutput(runner, []parser.FunctionListEntry{pick}, cpuPath, dir, nil); outErr != nil {
		t.Fatal(outErr)
	}
	outFile := filepath.Join(dir, pick.OutputStem+"."+workspace.TextExtension)
	if _, statErr := os.Stat(outFile); statErr != nil {
		t.Fatalf("expected output file: %v", statErr)
	}
}

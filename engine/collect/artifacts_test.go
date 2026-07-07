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

func TestRunPprofReport_fakeRunner(t *testing.T) {
	t.Parallel()
	runner := &tooling.FakeRunner{
		Out: [][]byte{[]byte("flat  flat%   sum%        cum   cum%")},
	}
	out := filepath.Join(t.TempDir(), "top.txt")
	if err := runPprofReport(runner, tooling.PprofTextReportArgs("top", "cpu.out"), out); err != nil {
		t.Fatal(err)
	}
	if len(runner.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runner.Runs))
	}
	if !strings.Contains(string(mustReadFile(t, out)), "flat") {
		t.Fatal("expected pprof text in output file")
	}
}

func TestRunPprofReport_nilRunner(t *testing.T) {
	t.Parallel()
	if err := runPprofReport(nil, tooling.PprofTextReportArgs("top", "cpu.out"), "out.txt"); err == nil {
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

func TestGetFunctionsOutput_parallelFakeRunner(t *testing.T) {
	t.Parallel()
	cpuPath := testpaths.MustAsset(t, "fixtures", filterFixtureCPU)
	entries, err := parser.GetFunctionListEntriesV2(cpuPath, config.FunctionFilter{})
	if err != nil {
		t.Fatal(err)
	}
	const want = 5
	if len(entries) < want {
		t.Fatalf("fixture should yield at least %d function entries, got %d", want, len(entries))
	}
	entries = entries[:want]

	outs := make([][]byte, len(entries))
	for i, e := range entries {
		outs[i] = []byte("list output for " + e.OutputStem)
	}
	runner := &tooling.FakeRunner{Out: outs}
	dir := t.TempDir()
	if outErr := getFunctionsOutput(runner, entries, cpuPath, dir, nil); outErr != nil {
		t.Fatal(outErr)
	}
	if len(runner.Runs) != len(entries) {
		t.Fatalf("expected %d pprof runs, got %d", len(entries), len(runner.Runs))
	}
	for _, e := range entries {
		outFile := filepath.Join(dir, e.OutputStem+"."+workspace.TextExtension)
		data, readErr := os.ReadFile(outFile)
		if readErr != nil {
			t.Fatalf("read %s: %v", outFile, readErr)
		}
		if len(data) == 0 {
			t.Fatalf("expected non-empty list output for %s", e.OutputStem)
		}
	}
}

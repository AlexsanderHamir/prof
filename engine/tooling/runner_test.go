package tooling

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestExecRunner_emptyArgv(t *testing.T) {
	r := NewExecRunner()
	_, err := r.Run(t.Context(), nil, RunOpts{})
	if err == nil || !strings.Contains(err.Error(), "empty argv") {
		t.Fatalf("expected empty argv error, got %v", err)
	}
}

func TestExecRunner_invalidCommand(t *testing.T) {
	r := NewExecRunner()
	argv := []string{"/nonexistent/binary-that-does-not-exist-xyz", "arg"}
	_, err := r.Run(t.Context(), argv, RunOpts{Combined: true})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecRunner_outputVsCombined(t *testing.T) {
	r := NewExecRunner()
	exe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	ctx := t.Context()

	out, err := r.Run(ctx, []string{exe, "env", "GOROOT"}, RunOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("expected stdout from go env GOROOT")
	}

	out2, err := r.Run(ctx, []string{exe, "env", "GOROOT"}, RunOpts{Combined: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(out2) == 0 {
		t.Fatal("expected combined output")
	}
}

func TestExecRunner_withDir(t *testing.T) {
	r := NewExecRunner()
	exe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	dir := t.TempDir()
	_, err = r.Run(t.Context(), []string{exe, "version"}, RunOpts{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecRunner_stdoutWriter(t *testing.T) {
	r := NewExecRunner()
	exe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	var buf bytes.Buffer
	_, err = r.Run(t.Context(), []string{exe, "version"}, RunOpts{Stdout: &buf})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "go version") {
		t.Fatalf("unexpected buffer: %q", buf.String())
	}
}

func TestFakeRunner_recordsAndReturns(t *testing.T) {
	f := &FakeRunner{
		Out: [][]byte{[]byte("a"), []byte("b")},
		Err: []error{nil, errors.New("boom")},
	}
	out1, err1 := f.Run(t.Context(), []string{"go", "version"}, RunOpts{Dir: t.TempDir()})
	if err1 != nil || string(out1) != "a" {
		t.Fatalf("first: out=%q err=%v", out1, err1)
	}
	if len(f.Runs) != 1 || len(f.Runs[0].Argv) != 2 {
		t.Fatalf("runs: %#v", f.Runs)
	}
	out2, err2 := f.Run(t.Context(), []string{"x"}, RunOpts{Combined: true})
	if err2 == nil || string(out2) != "b" {
		t.Fatalf("second: out=%q err=%v", out2, err2)
	}
}

func TestFakeRunner_nilReceiver(t *testing.T) {
	var f *FakeRunner
	_, err := f.Run(t.Context(), []string{"a"}, RunOpts{})
	if err == nil || !strings.Contains(err.Error(), "nil FakeRunner") {
		t.Fatalf("expected nil runner error, got %v", err)
	}
}

func TestDefaultCatalog(t *testing.T) {
	c := DefaultCatalog()
	ids := c.ProfileIDsSorted()
	if len(ids) != 4 {
		t.Fatalf("ids: %v", ids)
	}
	for _, id := range []string{"cpu", "memory", "mutex", "block"} {
		if err := c.ValidateProfile(id); err != nil {
			t.Fatal(err)
		}
	}
	if c.ValidateProfile("nope") == nil {
		t.Fatal("expected error")
	}
	args, err := c.GoTestProfileArgs([]string{"cpu", "memory"})
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != "-cpuprofile=cpu.out" || args[1] != "-memprofile=memory.out" {
		t.Fatalf("args: %v", args)
	}
	if _, ok := c.OutFileName("cpu"); !ok {
		t.Fatal("expected cpu.out mapping")
	}
	if name, ok := c.OutFileName("cpu"); !ok || name != "cpu.out" {
		t.Fatalf("got %q %v", name, ok)
	}
	m := c.KnownProfileSet()
	if len(m) != 4 {
		t.Fatalf("set size %d", len(m))
	}
}

func TestCatalog_nil(t *testing.T) {
	var c *Catalog
	if c.ProfileIDs() != nil {
		t.Fatal("expected nil")
	}
	if err := c.ValidateProfile("cpu"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := c.GoTestProfileArgs([]string{"cpu"}); err == nil {
		t.Fatal("expected error")
	}
}

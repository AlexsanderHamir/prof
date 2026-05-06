package cursoragent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClient_Run_fakeStreamExec(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "cursor-agent-fake")
	if err := os.WriteFile(bin, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	wantDir := dir
	resultLine := `{"type":"result","subtype":"success","result":"analysis done"}`
	fake := func(ctx context.Context, dir string, env []string, stdin []byte, name string, onStdoutLine func([]byte), args ...string) ([]byte, []byte, int, error) {
		if dir != wantDir {
			t.Fatalf("dir=%q want %q", dir, wantDir)
		}
		if !strings.Contains(string(stdin), "prompt body") {
			t.Fatalf("stdin=%q", stdin)
		}
		if onStdoutLine != nil {
			onStdoutLine([]byte(`{"type":"system","subtype":"init"}`))
			onStdoutLine([]byte(resultLine))
		}
		return []byte(resultLine + "\n"), nil, 0, nil
	}
	c := NewClient(Options{BinaryPath: bin, StreamExec: fake})
	req := RunRequest{
		Prompt:     []byte("prompt body"),
		WorkingDir: dir,
	}
	res, err := c.Run(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "analysis done" || res.ExitCode != 0 {
		t.Fatalf("got %+v err=%v", res, err)
	}
}

func TestClient_Run_emptyPrompt(t *testing.T) {
	c := NewClient(Options{StreamExec: func(context.Context, string, []string, []byte, string, func([]byte), ...string) ([]byte, []byte, int, error) {
		return nil, nil, 0, nil
	}})
	_, err := c.Run(t.Context(), RunRequest{WorkingDir: t.TempDir(), Prompt: []byte("  ")})
	if err == nil || !strings.Contains(err.Error(), "empty Prompt") {
		t.Fatalf("got %v", err)
	}
}

func TestClient_Run_badWorkingDir(t *testing.T) {
	c := NewClient(Options{StreamExec: func(context.Context, string, []string, []byte, string, func([]byte), ...string) ([]byte, []byte, int, error) {
		return nil, nil, 0, nil
	}})
	_, err := c.Run(t.Context(), RunRequest{WorkingDir: "/nonexistent/dir/xyz123", Prompt: []byte("x")})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_Probe_fake(t *testing.T) {
	fakeProbe := func(ctx context.Context, name string, args ...string) ([]byte, []byte, int, error) {
		if len(args) != 1 || args[0] != "--version" {
			t.Fatalf("args=%v", args)
		}
		return []byte("9.9.9-test\n"), nil, 0, nil
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "fake-agent")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := NewClient(Options{
		BinaryPath: p,
		ProbeFn:    fakeProbe,
	})
	resolved, ver, err := c.Probe(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if ver != "9.9.9-test" {
		t.Fatalf("version %q resolved %q", ver, resolved)
	}
	if resolved == "" {
		t.Fatal("empty resolved")
	}
}

func TestClient_Run_nonZeroExit(t *testing.T) {
	dir := t.TempDir()
	fake := func(context.Context, string, []string, []byte, string, func([]byte), ...string) ([]byte, []byte, int, error) {
		return nil, []byte("boom"), 7, nil
	}
	c := NewClient(Options{BinaryPath: filepath.Join(dir, "x"), StreamExec: fake})
	_ = os.WriteFile(filepath.Join(dir, "x"), []byte("x"), 0o644)

	_, err := c.Run(t.Context(), RunRequest{
		Prompt:     []byte("p"),
		WorkingDir: dir,
	})
	if err == nil || !errors.Is(err, ErrNonZeroExit) {
		t.Fatalf("got %v", err)
	}
}

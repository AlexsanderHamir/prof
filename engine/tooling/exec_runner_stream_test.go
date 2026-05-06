package tooling

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunWithStdinStreamStdout_emptyArgv(t *testing.T) {
	_, _, _, err := RunWithStdinStreamStdout(t.Context(), nil, StreamRunOpts{})
	if err == nil || !strings.Contains(err.Error(), "empty argv") {
		t.Fatalf("expected empty argv error, got %v", err)
	}
}

func TestRunWithStdinStreamStdout_goVersion(t *testing.T) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	var lines int
	stdout, stderr, code, err := RunWithStdinStreamStdout(t.Context(), []string{goExe, "version"}, StreamRunOpts{
		OnStdoutLine: func([]byte) { lines++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr)
	}
	if !strings.Contains(string(stdout), "go version") {
		t.Fatalf("stdout=%q", stdout)
	}
	if lines < 1 {
		t.Fatalf("expected at least one stdout line callback, got %d", lines)
	}
}

func TestRunWithStdinStreamStdout_multiLineProgram(t *testing.T) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	dir := t.TempDir()
	if wfErr := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module streamtest\n\ngo 1.24\n"), 0o600); wfErr != nil {
		t.Fatal(wfErr)
	}
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("L1")
	fmt.Println("L2")
}
`
	if wfErr := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0o600); wfErr != nil {
		t.Fatal(wfErr)
	}

	var got []string
	stdout, stderr, code, err := RunWithStdinStreamStdout(t.Context(), []string{goExe, "run", "."}, StreamRunOpts{
		Dir: dir,
		Env: minimalGoEnv(t),
		OnStdoutLine: func(line []byte) {
			got = append(got, string(line))
		},
	})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	if code != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if len(got) < 2 {
		t.Fatalf("expected 2+ line callbacks, got %d: %q", len(got), got)
	}
	joined := strings.Join(got, "|")
	if !strings.Contains(joined, "L1") || !strings.Contains(joined, "L2") {
		t.Fatalf("lines=%q stdout=%q", got, stdout)
	}
}

func TestRunWithStdinStreamStdout_stdin(t *testing.T) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	dir := t.TempDir()
	if wfErr := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module stdinmod\n\ngo 1.24\n"), 0o600); wfErr != nil {
		t.Fatal(wfErr)
	}
	mainGo := `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		fmt.Println(s.Text())
	}
}
`
	if wfErr := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0o600); wfErr != nil {
		t.Fatal(wfErr)
	}
	stdin := []byte("hello\n")
	stdout, stderr, code, err := RunWithStdinStreamStdout(t.Context(), []string{goExe, "run", "."}, StreamRunOpts{
		Dir:   dir,
		Env:   minimalGoEnv(t),
		Stdin: stdin,
	})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr)
	}
	if !strings.Contains(string(stdout), "hello") {
		t.Fatalf("stdout=%q", stdout)
	}
}

// minimalGoEnv returns an env suitable for nested `go run` in a temp module (avoids inheriting GOWORK from parent).
func minimalGoEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"USERPROFILE=" + os.Getenv("USERPROFILE"),
		"GOROOT=" + os.Getenv("GOROOT"),
		"GOPATH=" + os.Getenv("GOPATH"),
		"GOCACHE=" + os.Getenv("GOCACHE"),
		"GOMODCACHE=" + os.Getenv("GOMODCACHE"),
		"GOWORK=off",
	}
	if runtime.GOOS == "windows" {
		for _, k := range []string{"SYSTEMROOT", "WINDIR", "TEMP", "TMP", "LOCALAPPDATA", "USERPROFILE", "PROGRAMFILES"} {
			if v := os.Getenv(k); v != "" {
				env = append(env, k+"="+v)
			}
		}
	}
	return env
}

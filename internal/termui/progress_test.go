package termui

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestFormatProgressLabel(t *testing.T) {
	t.Parallel()

	got := formatProgressLabel(Progress{
		Label:  "BenchmarkFibonacci",
		Index:  2,
		Total:  2,
		Detail: "count=5",
	})
	want := "Running benchmark 2/2: BenchmarkFibonacci (count=5)…"
	if got != want {
		t.Fatalf("formatProgressLabel() = %q, want %q", got, want)
	}

	single := formatProgressLabel(Progress{Label: "BenchmarkOnly", Total: 1, Index: 1})
	if !strings.Contains(single, "Running benchmark: BenchmarkOnly") {
		t.Fatalf("single benchmark label = %q", single)
	}
}

func TestRunWhile_nonTTY(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	var calls int
	runErr := errors.New("boom")
	err = RunWhile(w, int(r.Fd()), Progress{Label: "B"}, func() error {
		calls++
		return runErr
	})
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if !errors.Is(err, runErr) {
		t.Fatalf("RunWhile() err = %v, want %v", err, runErr)
	}
}

func TestRunWhile_nilTask(t *testing.T) {
	t.Parallel()

	err := RunWhile(os.Stderr, int(os.Stderr.Fd()), Progress{}, nil)
	if err == nil {
		t.Fatal("expected error for nil task")
	}
}

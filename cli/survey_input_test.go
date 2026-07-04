package cli

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestAskConfigureLine_usesDefault(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	got, err := askConfigureLine(bufio.NewReader(strings.NewReader("\n")), &out, "Number of runs (count):", "1")
	if err != nil {
		t.Fatalf("askConfigureLine() err = %v", err)
	}
	if got != "1" {
		t.Fatalf("got %q, want 1", got)
	}
	if !strings.Contains(out.String(), "Number of runs (count):") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), configureQuestionIcon.Render("?")) {
		t.Fatalf("missing styled question icon: %q", out.String())
	}
}

func TestAskConfigureLine_required(t *testing.T) {
	t.Parallel()

	_, err := askConfigureLine(bufio.NewReader(strings.NewReader("\n")), io.Discard, "Tag name:", "")
	if err == nil {
		t.Fatal("expected required error")
	}
}

package cursoragent

import (
	"strings"
	"testing"
)

func TestParseStdout_singleResultJSON(t *testing.T) {
	in := `{"type":"result","subtype":"success","result":"hello"}`
	out, err := parseStdout([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if out.Result != "hello" || out.Type != eventResult {
		t.Fatalf("got %+v", out)
	}
}

func TestParseStdout_ndjsonAssistantFallback(t *testing.T) {
	in := `
{"type":"system","subtype":"init","model":"m1"}
{"type":"assistant","message":{"content":[{"type":"text","text":"partial"}]}}
`
	out, err := parseStdout([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if out.Result != "partial" {
		t.Fatalf("want partial, got %+v", out)
	}
	if !out.MissingTerminalResult {
		t.Fatal("expected MissingTerminalResult")
	}
	if out.ResolvedModel != "m1" {
		t.Fatalf("model %q", out.ResolvedModel)
	}
}

func TestParseStdout_unknownLinesIgnored(t *testing.T) {
	in := `{"foo":1}
{"type":"assistant","message":{"content":[{"type":"text","text":"ok"}]}}
`
	out, err := parseStdout([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if out.Result != "ok" {
		t.Fatalf("got %+v", out)
	}
}

func TestParseStdout_unknownTypeLineSkipped(t *testing.T) {
	in := `{"type":"noise","x":1}
{"type":"assistant","message":{"content":[{"type":"text","text":"ok"}]}}
`
	out, err := parseStdout([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if out.Result != "ok" {
		t.Fatalf("got %+v", out)
	}
}

func TestParseStdout_empty(t *testing.T) {
	_, err := parseStdout([]byte("   "))
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("got %v", err)
	}
}

func TestMergeBinaryPath(t *testing.T) {
	if got := MergeBinaryPath(" /a ", ""); got != "/a" {
		t.Fatalf("flag: %q", got)
	}
	if got := MergeBinaryPath("", " /b "); got != "/b" {
		t.Fatalf("env: %q", got)
	}
	if got := MergeBinaryPath("", ""); got != "" {
		t.Fatalf("empty: %q", got)
	}
}

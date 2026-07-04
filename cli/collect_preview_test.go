package cli

import (
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/termui"
)

func TestFormatWarningLine_matchesPipelineStyle(t *testing.T) {
	t.Parallel()

	got := termui.FormatWarningLine(termui.ConfigureDetailPrefix, "test warn")
	if !strings.Contains(got, "warning:") {
		t.Fatalf("missing warning prefix: %q", got)
	}
	if !strings.Contains(got, "test warn") {
		t.Fatalf("missing message: %q", got)
	}
}

func TestFormatFilterList(t *testing.T) {
	t.Parallel()
	if got := formatFilterList(nil); got != "(all)" {
		t.Fatalf("got %q", got)
	}
	if got := formatFilterList([]string{"a", "b"}); got != "a, b" {
		t.Fatalf("got %q", got)
	}
}

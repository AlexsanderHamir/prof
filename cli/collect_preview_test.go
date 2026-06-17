package cli

import "testing"

func TestFormatFilterList(t *testing.T) {
	t.Parallel()
	if got := formatFilterList(nil); got != "(all)" {
		t.Fatalf("got %q", got)
	}
	if got := formatFilterList([]string{"a", "b"}); got != "a, b" {
		t.Fatalf("got %q", got)
	}
}

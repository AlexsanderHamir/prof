package collector

import "testing"

func TestStemFromPath(t *testing.T) {
	if got := stemFromPath("/x/y/cpu.out"); got != "cpu" {
		t.Fatalf("stemFromPath: got %q", got)
	}
	if got := stemFromPath(`C:\a\b\mem.out`); got != "mem" {
		t.Fatalf("stemFromPath windows-ish: got %q", got)
	}
}

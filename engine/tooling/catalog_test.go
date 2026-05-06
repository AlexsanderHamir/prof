package tooling

import "testing"

func TestGoTestProfileArgs_unknown(t *testing.T) {
	c := DefaultCatalog()
	_, err := c.GoTestProfileArgs([]string{"cpu", "trace"})
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestNormalizeProfileCSV(t *testing.T) {
	got := NormalizeProfileCSV(" cpu , memory ")
	if len(got) != 2 || got[0] != "cpu" || got[1] != "memory" {
		t.Fatalf("%#v", got)
	}
	if NormalizeProfileCSV("") != nil {
		t.Fatal("expected nil")
	}
}

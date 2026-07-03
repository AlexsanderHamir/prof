package tooling

import (
	"slices"
	"testing"
)

func TestPprofTextTopArgs(t *testing.T) {
	got := PprofTextTopArgs("/tmp/cpu.out")
	want := []string{"go", "tool", "pprof", "-cum", "-edgefraction=0", "-nodefraction=0", "-top", "/tmp/cpu.out"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestPprofPNGArgs(t *testing.T) {
	got := PprofPNGArgs("b.out")
	if got[len(got)-1] != "b.out" || got[3] != "-png" {
		t.Fatalf("got %v", got)
	}
}

func TestPprofListArgs(t *testing.T) {
	got := PprofListArgs("b.out", `main\.foo`)
	if got[3] != `-list=main\.foo` || got[4] != "b.out" {
		t.Fatalf("got %v", got)
	}
}

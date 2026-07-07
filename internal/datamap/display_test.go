package datamap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/pprofscale"
	"github.com/AlexsanderHamir/prof/parser"
)

func cpuFixturePath(t *testing.T) string {
	t.Helper()
	for _, rel := range []string{
		"tests/assets/fixtures/BenchmarkStringProcessor_cpu.out",
		filepath.Join("..", "..", "tests", "assets", "fixtures", "BenchmarkStringProcessor_cpu.out"),
	} {
		if _, err := os.Stat(rel); err == nil {
			return rel
		}
	}
	t.Skip("cpu profile fixture not present")
	return ""
}

// Display helpers must match go tool pprof -top for profile totals (used on profiles section).
func TestPprofDisplay_matchesPprofTop(t *testing.T) {
	t.Parallel()
	path := cpuFixturePath(t)
	d, err := parser.DefaultPipeline().RunFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if d.SampleUnit != "nanoseconds" {
		t.Fatalf("sample_unit=%q want nanoseconds", d.SampleUnit)
	}

	want := map[string]string{
		"crypto/internal/fips140/sha256.blockSHANI": "0.19s",
		"runtime.memmove":                           "0.16s",
	}
	for sym, label := range want {
		flat, ok := d.Flat[sym]
		if !ok {
			t.Fatalf("symbol %q missing from profile", sym)
		}
		outUnit := pprofscale.SelectOutputUnit(d.SampleUnit, d.Total, d.Flat, d.Cum)
		got := pprofscale.ScaledLabel(flat, d.SampleUnit, outUnit)
		if got != label {
			t.Fatalf("%s flat display=%q want %q (raw=%d outUnit=%q)", sym, got, label, flat, outUnit)
		}
	}

	display, sec, outUnit := profileTotalDisplay(d)
	if display != "3.15s" || sec != 3.15 || outUnit != "s" {
		t.Fatalf("total display=%q seconds=%v outUnit=%q want 3.15s/s", display, sec, outUnit)
	}
}

package pprofscale

import "testing"

func TestScaledLabel_cpuNanoseconds(t *testing.T) {
	t.Parallel()
	got := ScaledLabel(1_470_000_000, "nanoseconds", "auto")
	if got != "1.47s" {
		t.Fatalf("got %q want 1.47s", got)
	}
}

func TestSeconds_cpuNanoseconds(t *testing.T) {
	t.Parallel()
	sec, ok := Seconds(1_470_000_000, "nanoseconds")
	if !ok {
		t.Fatal("expected ok")
	}
	if sec != 1.47 {
		t.Fatalf("got %v want 1.47", sec)
	}
}

func TestScaledLabel_memoryBytes(t *testing.T) {
	t.Parallel()
	const bytes = int64(891_425_914) // 850.13 * 2^20
	got := ScaledLabel(bytes, "bytes", "auto")
	if got != "850.13MB" {
		t.Fatalf("got %q want 850.13MB", got)
	}
}

func TestSelectOutputUnit_cpuProfile(t *testing.T) {
	t.Parallel()
	flat := map[string]int64{"a": 10_000_000}
	cum := map[string]int64{"a": 10_000_000}
	const total = 3_150_000_000
	got := SelectOutputUnit("nanoseconds", total, flat, cum)
	if got != "s" {
		t.Fatalf("got %q want s", got)
	}
	if ScaledLabel(190_000_000, "nanoseconds", got) != "0.19s" {
		t.Fatal("expected 0.19s with profile output unit")
	}
}

func TestScale_totalCPUProfile(t *testing.T) {
	t.Parallel()
	if ScaledLabel(10_220_000_000, "nanoseconds", "auto") != "10.22s" {
		t.Fatal("total label mismatch")
	}
	sec, ok := Seconds(10_220_000_000, "nanoseconds")
	if !ok || sec != 10.22 {
		t.Fatalf("seconds=%v ok=%v", sec, ok)
	}
}

package benchmark

import "testing"

func TestBuildBenchmarkCommand_CPUAndMemory(t *testing.T) {
	cmd, err := buildBenchmarkCommand("BenchmarkFoo", []string{"cpu", "memory"}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd) < 6 {
		t.Fatalf("cmd too short: %v", cmd)
	}
	if cmd[0] != "go" || cmd[1] != "test" {
		t.Fatalf("unexpected argv: %v", cmd)
	}
}

func TestBuildBenchmarkCommandUnsupportedProfile(t *testing.T) {
	_, err := buildBenchmarkCommand("BenchmarkFoo", []string{"not-a-profile"}, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetExpectedProfileFileName(t *testing.T) {
	name, ok := getExpectedProfileFileName("cpu")
	if !ok || name != "cpu.out" {
		t.Fatalf("got %q %v", name, ok)
	}
	_, ok = getExpectedProfileFileName("nope")
	if ok {
		t.Fatal("expected false")
	}
}

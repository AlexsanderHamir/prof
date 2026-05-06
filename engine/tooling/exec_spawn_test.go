package tooling

import "testing"

func TestLookPath_go(t *testing.T) {
	_, err := LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
}

func TestStartDetached_emptyArgv(t *testing.T) {
	err := StartDetached(t.Context(), nil, RunOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStartDetached_rejectsMissingBinary(t *testing.T) {
	err := StartDetached(t.Context(), []string{"/nonexistent/qcachegrind-tooling-test-xyz"}, RunOpts{})
	if err == nil {
		t.Fatal("expected error starting missing binary")
	}
}

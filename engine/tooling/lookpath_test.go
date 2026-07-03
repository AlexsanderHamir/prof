package tooling

import "testing"

func TestLookPath_go(t *testing.T) {
	_, err := LookPath("go")
	if err != nil {
		t.Fatalf("LookPath(go): %v", err)
	}
}

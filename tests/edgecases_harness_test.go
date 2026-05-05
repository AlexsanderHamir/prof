package tests

import (
	"os"
	"testing"
)

// TestEdge_fixtureHarness is a smoke check that edge-case helpers resolve the
// same committed fixtures as the rest of the integration suite.
func TestEdge_fixtureHarness(t *testing.T) {
	t.Run("cpu_fixture_present", func(t *testing.T) {
		p := edgecasesFixturePath(t, fixtureCPUFile)
		st, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
		if st.Size() == 0 {
			t.Fatalf("empty fixture: %s", p)
		}
	})
	t.Run("memory_fixture_present", func(t *testing.T) {
		p := edgecasesFixturePath(t, fixtureMemFile)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
	})
}

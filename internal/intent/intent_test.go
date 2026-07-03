package intent

import "testing"

func TestAllKinds(t *testing.T) {
	got := AllKinds()
	if len(got) != 5 {
		t.Fatalf("AllKinds: want 5 entries, got %d", len(got))
	}
	seen := map[Kind]bool{}
	for _, d := range got {
		if d.Description == "" {
			t.Errorf("empty description for kind %q", d.K)
		}
		seen[d.K] = true
	}
	for _, k := range []Kind{KindCollect, KindCompare, KindSetup, KindConfigCreate, KindConfigSave} {
		if !seen[k] {
			t.Errorf("missing kind %q in AllKinds", k)
		}
	}
}

package collector

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

func TestWriteGroupedPackageProfileMissingBinary(t *testing.T) {
	out := filepath.Join(t.TempDir(), "g.txt")
	err := WriteGroupedPackageProfile(filepath.Join(t.TempDir(), "missing.out"), out, internal.FunctionFilter{})
	if err == nil {
		t.Fatal("expected error for missing profile binary")
	}
}

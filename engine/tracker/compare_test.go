package tracker

import (
	"os"
	"testing"
)

func TestLoadProfileObjects_missingBaseline(t *testing.T) {
	t.Parallel()
	sel := &Options{
		Baseline:      "baseline-v1",
		Current:       "candidate-v2",
		BenchmarkName: "BenchmarkStringProcessor",
		ProfileType:   "cpu",
	}
	_, err := loadProfileObjects(filepathJoinMissing(), sel, "baseline")
	if err == nil {
		t.Fatal("expected error")
	}
	want := `baseline tag "baseline-v1" has no BenchmarkStringProcessor/cpu profile — run collect for that benchmark or pick another`
	if got := err.Error(); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func filepathJoinMissing() string {
	return os.TempDir() + string(os.PathSeparator) + "prof-missing-profile-test.out"
}

package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"

	pprofprofile "github.com/google/pprof/profile"
)

func benchmarkGenPoolCPUFixturePath(t *testing.T) string {
	t.Helper()
	p := filepath.Join("testdata", "testFilesV2", "BenchmarkGenPool_cpu.out")
	if _, err := os.Stat(p); err != nil {
		t.Skip("fixture not present:", p)
	}
	return p
}

func TestProfileDataFromPath_CPUFixture(t *testing.T) {
	path := benchmarkGenPoolCPUFixturePath(t)
	d, err := profileDataFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if d.Total <= 0 || len(d.SortedEntries) == 0 {
		t.Fatalf("unexpected empty aggregation: total=%d entries=%d", d.Total, len(d.SortedEntries))
	}
}

func TestDefaultPipelineRunFromPath(t *testing.T) {
	path := benchmarkGenPoolCPUFixturePath(t)
	p := DefaultPipeline()
	d, err := p.RunFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if d == nil || len(d.Flat) == 0 {
		t.Fatal("expected non-empty ProfileData")
	}
}

func TestPipelineWithDefaultsPartialOverride(t *testing.T) {
	pl := Pipeline{Validator: StandardProfileValidator{}}
	pl = pl.withDefaults()
	if pl.Opener == nil || pl.Decoder == nil {
		t.Fatal("withDefaults should fill nil fields")
	}
}

func TestValidateProfileNil(t *testing.T) {
	if err := ValidateProfile(nil); err == nil {
		t.Fatal("expected error for nil profile")
	}
}

func TestPrimarySampleValueIndexEmptyTypes(t *testing.T) {
	p := &pprofprofile.Profile{Sample: []*pprofprofile.Sample{{Value: []int64{1}}}}
	_, err := PrimarySampleValueIndex(p)
	if err == nil {
		t.Fatal("expected error when SampleType empty")
	}
}

func TestGetAllFunctionNamesFromProfileDataFilters(t *testing.T) {
	d := &ProfileData{
		SortedEntries: []FuncEntry{
			{Name: "example.com/pkg.(*T).Method", Flat: 10},
			{Name: "other.Short", Flat: 5},
		},
	}
	all := GetAllFunctionNamesFromProfileData(d, config.FunctionFilter{})
	if len(all) != 2 {
		t.Fatalf("got %v", all)
	}
	ign := GetAllFunctionNamesFromProfileData(d, config.FunctionFilter{IgnoreFunctions: []string{"Method"}})
	if len(ign) != 1 || ign[0] != "Short" {
		t.Fatalf("got %v", ign)
	}
	pref := GetAllFunctionNamesFromProfileData(d, config.FunctionFilter{IncludePrefixes: []string{"example.com"}})
	if len(pref) != 1 {
		t.Fatalf("got %v", pref)
	}
}

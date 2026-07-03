package parser

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"

	pprofprofile "github.com/google/pprof/profile"
)

func TestLoadProfileAndParseWrappers(t *testing.T) {
	if _, err := LoadProfile(filepath.Join(t.TempDir(), "nonexistent.out")); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ParseProfileFromReader(strings.NewReader("not protobuf profile data")); err == nil {
		t.Fatal("expected parse error")
	}
	path := benchmarkGenPoolCPUFixturePath(t)
	if _, err := ParseProfileFromPath(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error")
	}
	p, err := ParseProfileFromPath(path)
	if err != nil || p == nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	pd, rErr := ProfileDataFromReader(f)
	if rErr != nil {
		t.Fatal(rErr)
	}
	_ = pd
}

func TestPathBasedFacadeFuncs(t *testing.T) {
	path := benchmarkGenPoolCPUFixturePath(t)
	if _, missingErr := GetAllFunctionNamesV2(filepath.Join(t.TempDir(), "nope"), config.FunctionFilter{}); missingErr == nil {
		t.Fatal()
	}
	names, err := GetAllFunctionNamesV2(path, config.FunctionFilter{})
	if err != nil || len(names) == 0 {
		t.Fatal(err, len(names))
	}
}

func TestValidateProfileCheckValidAndSampleTypes(t *testing.T) {
	p := &pprofprofile.Profile{}
	if err := ValidateProfile(p); err == nil {
		t.Fatal("expected CheckValid or sample type error")
	}
	ok, err := pprofprofile.Parse(strings.NewReader(""))
	if err == nil && ok != nil {
		_ = ValidateProfile(ok)
	}
	path := benchmarkGenPoolCPUFixturePath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := pprofprofile.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	p2.SampleType = nil
	if validateErr := ValidateProfile(p2); validateErr == nil || !strings.Contains(validateErr.Error(), "sample") {
		t.Fatalf("got %v", validateErr)
	}
}

func TestValidateSamplesHaveValueAtBranches(t *testing.T) {
	if err := ValidateSamplesHaveValueAt(&pprofprofile.Profile{}, -1); err == nil {
		t.Fatal()
	}
	p := &pprofprofile.Profile{
		Sample: []*pprofprofile.Sample{nil},
	}
	if err := ValidateSamplesHaveValueAt(p, 0); err == nil {
		t.Fatal()
	}
	p.Sample = []*pprofprofile.Sample{{Value: []int64{}}}
	if err := ValidateSamplesHaveValueAt(p, 0); err == nil {
		t.Fatal()
	}
}

func TestAccumulateSampleBranches(t *testing.T) {
	flat, cum := make(map[string]int64), make(map[string]int64)
	fn := &pprofprofile.Function{Name: "F"}
	// Flat counts only the first line of the top stack frame; put F first so both flat and cum see it.
	loc := &pprofprofile.Location{Line: []pprofprofile.Line{{Function: fn}, {Function: nil}}}
	s := &pprofprofile.Sample{
		Location: []*pprofprofile.Location{loc},
		Value:    []int64{3},
	}
	accumulateSample(s, 3, flat, cum)
	if cum["F"] != 3 || flat["F"] != 3 {
		t.Fatalf("flat=%v cum=%v", flat, cum)
	}
	flat2, cum2 := make(map[string]int64), make(map[string]int64)
	accumulateSample(&pprofprofile.Sample{Location: nil, Value: []int64{1}}, 1, flat2, cum2)
	if len(flat2) != 0 || len(cum2) != 0 {
		t.Fatal()
	}
	topNoFn := &pprofprofile.Location{Line: []pprofprofile.Line{{Function: nil}}}
	flat3, cum3 := make(map[string]int64), make(map[string]int64)
	accumulateSample(&pprofprofile.Sample{Location: []*pprofprofile.Location{topNoFn}, Value: []int64{1}}, 1, flat3, cum3)
	if len(flat3) != 0 {
		t.Fatal()
	}
}

func TestAggregateEmptySamples(_ *testing.T) {
	p := &pprofprofile.Profile{SampleType: []*pprofprofile.ValueType{{Type: "t", Unit: "u"}}}
	_ = AggregateProfileData(p, 0)
}

func TestSimpleFunctionNameTable(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"single", "single"},
		{"a.b.Foo", "Foo"},
		{"pkg.(*T).Method", "Method"},
		{"x.Type[U].Method", "Method"},
		{"x.Func()", "Func"},
	}
	for _, tc := range cases {
		if got := simpleFunctionName(tc.in); got != tc.want {
			t.Errorf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

type errDecoder struct{}

func (errDecoder) Decode(io.Reader) (*pprofprofile.Profile, error) { return nil, errors.New("dec") }

// fixedDecode returns the same profile for any reader (for pipeline error-path tests after decode).
type fixedDecodeProfile struct{ p *pprofprofile.Profile }

func (f fixedDecodeProfile) Decode(io.Reader) (*pprofprofile.Profile, error) { return f.p, nil }

type errValidator struct{}

func (errValidator) Validate(*pprofprofile.Profile) error { return errors.New("bad") }

type errIndex struct{}

func (errIndex) PrimaryIndex(*pprofprofile.Profile) (int, error) { return 0, errors.New("idx") }

type errShape struct{}

func (errShape) EnsureValueAt(*pprofprofile.Profile, int) error { return errors.New("shape") }

func TestPipelineRunFromReaderErrors(t *testing.T) {
	_, err := Pipeline{Decoder: errDecoder{}}.RunFromReader(strings.NewReader("x"))
	if err == nil {
		t.Fatal()
	}
	data, err := os.ReadFile(benchmarkGenPoolCPUFixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	p, err := pprofprofile.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	dec := fixedDecodeProfile{p: p}
	_, err = Pipeline{Decoder: dec, Validator: errValidator{}}.RunFromReader(strings.NewReader("x"))
	if err == nil {
		t.Fatal()
	}
	_, err = Pipeline{
		Decoder: dec, Validator: StandardProfileValidator{}, IndexSelect: errIndex{},
	}.RunFromReader(strings.NewReader("x"))
	if err == nil {
		t.Fatal()
	}
	_, err = Pipeline{
		Decoder: dec, Validator: StandardProfileValidator{},
		IndexSelect: FirstSampleIndexSelector{}, IndexCheck: errShape{},
	}.RunFromReader(strings.NewReader("x"))
	if err == nil {
		t.Fatal()
	}
}

type errOpener struct{}

func (errOpener) Open(string) (io.ReadCloser, error) { return nil, errors.New("open") }

func TestPipelineRunFromPathOpenError(t *testing.T) {
	_, err := Pipeline{Opener: errOpener{}}.RunFromPath("any")
	if err == nil {
		t.Fatal()
	}
}

func TestFileProfileOpenerMissing(t *testing.T) {
	_, err := FileProfileOpener{}.Open(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal()
	}
}

func TestPProfDecoderInvalid(t *testing.T) {
	_, err := PProfDecoder{}.Decode(strings.NewReader("@@@"))
	if err == nil {
		t.Fatal()
	}
}

func TestWithDefaultsEachSlot(t *testing.T) {
	// Ensure every nil branch in withDefaults can be hit via a pipeline that only sets one field to non-default.
	pl := Pipeline{}
	pl = pl.withDefaults()
	if pl.Opener == nil {
		t.Fatal()
	}
	pl2 := Pipeline{Opener: FileProfileOpener{}}
	pl2 = pl2.withDefaults()
	if pl2.Decoder == nil {
		t.Fatal()
	}
}

func TestGetAllFunctionNamesFromProfileDataEmptyShort(t *testing.T) {
	d := &ProfileData{SortedEntries: []FuncEntry{{Name: ".", Flat: 1}}}
	n := GetAllFunctionNamesFromProfileData(d, config.FunctionFilter{})
	if len(n) != 0 {
		t.Fatal(n)
	}
}

func TestGetAllFunctionNamesFromProfileDataNilAndFilters(t *testing.T) {
	if GetAllFunctionNamesFromProfileData(nil, config.FunctionFilter{}) != nil {
		t.Fatal()
	}
	d := &ProfileData{
		SortedEntries: []FuncEntry{
			{Name: "pkg.A", Flat: 1},
			{Name: "pkg.B", Flat: 1},
			{Name: "other.C", Flat: 1},
		},
	}
	if n := GetAllFunctionNamesFromProfileData(d, config.FunctionFilter{IgnoreFunctions: []string{"A"}}); len(n) != 2 {
		t.Fatal(n)
	}
	if n := GetAllFunctionNamesFromProfileData(d, config.FunctionFilter{IncludePrefixes: []string{"other"}}); len(n) != 1 || n[0] != "C" {
		t.Fatal(n)
	}
}

package parser

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
	pprofprofile "github.com/google/pprof/profile"
)

func TestLoadProfileAndParseWrappers(t *testing.T) {
	if _, err := LoadProfile(filepath.Join(t.TempDir(), "nonexistent.out")); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ParseProfileFromReader(strings.NewReader("not protobuf profile data")); err == nil {
		t.Fatal("expected parse error")
	}
	path := testProfilePath(t, "BenchmarkGenPool_cpu.out")
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
	if _, err := ProfileDataFromReader(f); err != nil {
		t.Fatal(err)
	}
}

func TestPathBasedFacadeFuncs(t *testing.T) {
	path := testProfilePath(t, "BenchmarkGenPool_cpu.out")
	if _, err := TurnLinesIntoObjectsV2(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error")
	}
	objs, err := TurnLinesIntoObjectsV2(path)
	if err != nil || len(objs) == 0 {
		t.Fatal(err, len(objs))
	}
	if _, err := GetAllFunctionNamesV2(filepath.Join(t.TempDir(), "nope"), internal.FunctionFilter{}); err == nil {
		t.Fatal()
	}
	names, err := GetAllFunctionNamesV2(path, internal.FunctionFilter{})
	if err != nil || len(names) == 0 {
		t.Fatal(err, len(names))
	}
	if _, err := OrganizeProfileByPackageV2(filepath.Join(t.TempDir(), "nope"), internal.FunctionFilter{}); err == nil {
		t.Fatal()
	}
	s, err := OrganizeProfileByPackageV2(path, internal.FunctionFilter{})
	if err != nil || !strings.Contains(s, "Subtotal") {
		t.Fatal(err, s)
	}
}

func TestLineObjsFromProfileDataNonNil(t *testing.T) {
	d := &ProfileData{
		SortedEntries: []FuncEntry{{Name: "a.B", Flat: 1}},
		FlatPercentages: map[string]float64{"a.B": 100},
		CumPercentages:  map[string]float64{"a.B": 100},
		SumPercentages:   map[string]float64{"a.B": 100},
		Cum:             map[string]int64{"a.B": 1},
	}
	objs := LineObjsFromProfileData(d)
	if len(objs) != 1 || objs[0].FnName != "a.B" {
		t.Fatalf("%+v", objs)
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
	path := testProfilePath(t, "BenchmarkGenPool_cpu.out")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := pprofprofile.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	p2.SampleType = nil
	if err := ValidateProfile(p2); err == nil || !strings.Contains(err.Error(), "sample") {
		t.Fatalf("got %v", err)
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

func TestAggregateEmptySamples(t *testing.T) {
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

func TestPackageNameFromSymbolTable(t *testing.T) {
	cases := []struct{ in, want string }{
		{"x", ""},
		{"fmt.Println", "fmt"},
		{"sync/atomic.Load", "sync/atomic"},
		{"github.com/a/b.Foo", "github.com/a/b"},
		{"golang.org/x/y.Z", "golang.org/x/y"},
		{"other/pkg.Sub", "other/pkg"},
	}
	for _, tc := range cases {
		if got := packageNameFromSymbol(tc.in); got != tc.want {
			t.Errorf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestShortPackageLabel(t *testing.T) {
	// One dot in the import path: split yields ["github", "com/foo/bar"]; label is the last segment.
	if got := shortPackageLabel("github.com/foo/bar"); got != "com/foo/bar" {
		t.Fatal(got)
	}
	if got := shortPackageLabel("github.com/foo.bar.baz"); got != "baz" {
		t.Fatal(got)
	}
	if got := shortPackageLabel("encoding/json"); got != "encoding/json" {
		t.Fatal(got)
	}
}

func TestOrganizeUnknownPackageFormatting(t *testing.T) {
	d := &ProfileData{
		Total: 10,
		SortedEntries: []FuncEntry{
			{Name: "orphanSymbolWithoutEnoughDots", Flat: 10},
		},
		FlatPercentages: map[string]float64{"orphanSymbolWithoutEnoughDots": 100},
		CumPercentages:  map[string]float64{"orphanSymbolWithoutEnoughDots": 100},
		SumPercentages:   map[string]float64{"orphanSymbolWithoutEnoughDots": 100},
		Cum:             map[string]int64{"orphanSymbolWithoutEnoughDots": 10},
	}
	s := OrganizeProfileByPackageFromProfileData(d, internal.FunctionFilter{})
	if !strings.Contains(s, "unknown") || !strings.Contains(s, "flat:") {
		t.Fatal(s)
	}
}

func TestSortPackagesByFlatMultiple(t *testing.T) {
	g := map[string]*PackageGroup{
		"a": {Name: "a", FlatPercentage: 10},
		"b": {Name: "b", FlatPercentage: 90},
	}
	sorted := sortPackagesByFlatPercentage(g)
	if len(sorted) != 2 || sorted[0].Name != "b" {
		t.Fatal(sorted)
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
	data, err := os.ReadFile(testProfilePath(t, "BenchmarkGenPool_cpu.out"))
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
	n := GetAllFunctionNamesFromProfileData(d, internal.FunctionFilter{})
	if len(n) != 0 {
		t.Fatal(n)
	}
}

func TestGetAllFunctionNamesFromProfileDataNilAndFilters(t *testing.T) {
	if GetAllFunctionNamesFromProfileData(nil, internal.FunctionFilter{}) != nil {
		t.Fatal()
	}
	d := &ProfileData{
		SortedEntries: []FuncEntry{
			{Name: "pkg.A", Flat: 1},
			{Name: "pkg.B", Flat: 1},
			{Name: "other.C", Flat: 1},
		},
	}
	if n := GetAllFunctionNamesFromProfileData(d, internal.FunctionFilter{IgnoreFunctions: []string{"A"}}); len(n) != 2 {
		t.Fatal(n)
	}
	if n := GetAllFunctionNamesFromProfileData(d, internal.FunctionFilter{IncludePrefixes: []string{"other"}}); len(n) != 1 || n[0] != "C" {
		t.Fatal(n)
	}
}

func TestOrganizeProfileByPackageFromProfileDataNilAndFilters(t *testing.T) {
	if OrganizeProfileByPackageFromProfileData(nil, internal.FunctionFilter{}) != "" {
		t.Fatal()
	}
	d := &ProfileData{
		Total:           2,
		SortedEntries:   []FuncEntry{{Name: "p.X", Flat: 1}, {Name: "p.Y", Flat: 1}},
		FlatPercentages: map[string]float64{"p.X": 50, "p.Y": 50},
		CumPercentages:  map[string]float64{"p.X": 1, "p.Y": 1},
		SumPercentages:   map[string]float64{"p.X": 50, "p.Y": 100},
		Cum:             map[string]int64{"p.X": 1, "p.Y": 1},
	}
	if s := OrganizeProfileByPackageFromProfileData(d, internal.FunctionFilter{IgnoreFunctions: []string{"X"}}); !strings.Contains(s, "Y") || strings.Contains(s, "X") {
		t.Fatal(s)
	}
	d2 := &ProfileData{
		Total:           1,
		SortedEntries:   []FuncEntry{{Name: "keep.M", Flat: 1}},
		FlatPercentages: map[string]float64{"keep.M": 100},
		CumPercentages:  map[string]float64{"keep.M": 1},
		SumPercentages:   map[string]float64{"keep.M": 100},
		Cum:             map[string]int64{"keep.M": 1},
	}
	if s := OrganizeProfileByPackageFromProfileData(d2, internal.FunctionFilter{IncludePrefixes: []string{"nomatch"}}); s != "" {
		t.Fatal(s)
	}
	d3 := &ProfileData{
		Total:           1,
		SortedEntries:   []FuncEntry{{Name: "NoDotSymbol", Flat: 1}},
		FlatPercentages: map[string]float64{"NoDotSymbol": 100},
		CumPercentages:  map[string]float64{"NoDotSymbol": 1},
		SumPercentages:   map[string]float64{"NoDotSymbol": 100},
		Cum:             map[string]int64{"NoDotSymbol": 1},
	}
	if s := OrganizeProfileByPackageFromProfileData(d3, internal.FunctionFilter{}); !strings.Contains(s, "unknown") {
		t.Fatal(s)
	}
	d4 := &ProfileData{
		Total:           1,
		SortedEntries:   []FuncEntry{{Name: ".", Flat: 1}},
		FlatPercentages: map[string]float64{".": 100},
		CumPercentages:  map[string]float64{".": 1},
		SumPercentages:   map[string]float64{".": 100},
		Cum:             map[string]int64{".": 1},
	}
	if s := OrganizeProfileByPackageFromProfileData(d4, internal.FunctionFilter{}); s != "" {
		t.Fatalf("expected empty report when all names filter to empty short, got %q", s)
	}
}

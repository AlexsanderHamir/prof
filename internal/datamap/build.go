package datamap

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

const (
	topSymbolLimit   = 10
	hotPathLimit     = 5
	statusOK         = "ok"
	statusSkipped    = "skipped"
	collectionAuto   = "auto"
	collectionManual = "manual"
)

var (
	defaultRecommendedFlow = []string{"measurements", "hotspots", "call_trees", "source_lines", "profiles"}
	defaultReadingGuide    = map[string]string{
		"measurements": "Go benchmark output (ns/op, B/op, allocs/op). Start here to confirm the run succeeded.",
		"hotspots":     "pprof -top: functions ranked by flat and cumulative cost.",
		"call_trees":   "pprof -tree: caller/callee context for top nodes.",
		"source_lines": "pprof -list extracts per function; use on symbols chosen from hotspots.",
		"profiles":     "Raw .out binaries; re-query with go tool pprof when text is insufficient.",
	}
)

// ProfileSnapshot captures in-memory state for one profile kind at map emit time.
type ProfileSnapshot struct {
	Profile              string
	ProfileData          *parser.ProfileData
	ListEntries          []parser.FunctionListEntry
	SourceLinesCollected int
	SourceLinesSkipped   int
	FailedStems          map[string]struct{}
}

// BuildInput is the collect → datamap contract.
type BuildInput struct {
	Layout           workspace.TagLayout
	Tag              string
	Benchmark        string
	Package          string
	CollectionMode   string
	Profiles         []string
	Filter           config.FunctionFilter
	BenchCount       int
	PerProfile       []ProfileSnapshot
	IncludeMeasuring bool
}

// Build assembles a BenchmarkMap from collect inputs without reading profile binaries again.
func Build(in BuildInput) (BenchmarkMap, error) {
	if in.Layout.Root == "" {
		return BenchmarkMap{}, errors.New("datamap: empty layout root")
	}
	if in.Benchmark == "" {
		return BenchmarkMap{}, errors.New("datamap: empty benchmark name")
	}

	m := BenchmarkMap{
		SchemaVersion:   SchemaVersion,
		Tag:             in.Tag,
		Benchmark:       in.Benchmark,
		Package:         in.Package,
		RecommendedFlow: append([]string(nil), defaultRecommendedFlow...),
		ReadingGuide:    copyReadingGuide(),
		Profiles:        make(map[string]ProfileRef, len(in.Profiles)),
		Hotspots:        make(map[string]HotspotSection, len(in.Profiles)),
		CallTrees:       make(map[string]CallTreeSection, len(in.Profiles)),
		SourceLines:     make(map[string]SourceLinesSection, len(in.Profiles)),
		CallGraphs:      make(map[string]CallGraphRef, len(in.Profiles)),
		Status: Status{
			Profiles:    make(map[string]string, len(in.Profiles)),
			Hotspots:    make(map[string]string, len(in.Profiles)),
			CallTrees:   make(map[string]string, len(in.Profiles)),
			CallGraphs:  make(map[string]CallGraphStatus, len(in.Profiles)),
			SourceLines: make(map[string]SourceLinesStatus, len(in.Profiles)),
		},
		Provenance: Provenance{
			Tag:               in.Tag,
			CollectionMode:    in.CollectionMode,
			BenchCount:        in.BenchCount,
			ProfilesRequested: append([]string(nil), in.Profiles...),
			Filter: FilterSnapshot{
				IncludePrefixes: append([]string(nil), in.Filter.IncludePrefixes...),
				IgnoreFunctions: append([]string(nil), in.Filter.IgnoreFunctions...),
			},
		},
	}

	if in.IncludeMeasuring && in.CollectionMode == collectionAuto {
		if err := m.addMeasurements(in); err != nil {
			return BenchmarkMap{}, err
		}
		m.Status.BenchmarkRun = statusOK
	}

	snapByProfile := make(map[string]ProfileSnapshot, len(in.PerProfile))
	for _, snap := range in.PerProfile {
		snapByProfile[snap.Profile] = snap
	}

	for _, profile := range in.Profiles {
		snap := snapByProfile[profile]
		if err := m.addProfileArtifacts(in, profile, snap); err != nil {
			return BenchmarkMap{}, err
		}
	}

	return m, nil
}

func copyReadingGuide() map[string]string {
	out := make(map[string]string, len(defaultReadingGuide))
	for k, v := range defaultReadingGuide {
		out[k] = v
	}
	return out
}

func (m *BenchmarkMap) addMeasurements(in BuildInput) error {
	abs := in.Layout.Measurement(in.Benchmark)
	rel, err := in.Layout.RelFromLayout(abs)
	if err != nil {
		return err
	}
	section := &MeasurementsSection{
		Path:        rel,
		Purpose:     PurposeBenchmemResults,
		Description: "Combined stdout from go test -bench with -benchmem.",
	}
	if summary, sumErr := parseMeasurementSummary(abs); sumErr == nil {
		section.Summary = summary
	}
	m.Measurements = section
	return nil
}

func (m *BenchmarkMap) addProfileArtifacts(in BuildInput, profile string, snap ProfileSnapshot) error {
	profRel, err := in.Layout.RelFromLayout(in.Layout.ProfileBinary(in.Benchmark, profile))
	if err != nil {
		return err
	}
	m.Profiles[profile] = ProfileRef{
		Path:         profRel,
		Purpose:      PurposeRawPprofBinary,
		Description:  "Raw pprof profile binary; source of truth for go tool pprof.",
		TotalSamples: snapTotal(snap.ProfileData),
	}
	m.Status.Profiles[profile] = statusOK

	hotRel, err := in.Layout.RelFromLayout(in.Layout.Hotspot(in.Benchmark, profile))
	if err != nil {
		return err
	}
	m.Hotspots[profile] = HotspotSection{
		Path:        hotRel,
		Purpose:     PurposeFlatCumulativeRanking,
		Description: "go tool pprof -top output: flat time in function body, cum time including callees.",
		Producer:    "go tool pprof -top",
		TopSymbols:  topSymbols(snap.ProfileData, topSymbolLimit),
	}
	m.Status.Hotspots[profile] = statusOK

	treeRel, err := in.Layout.RelFromLayout(in.Layout.CallTreeText(in.Benchmark, profile))
	if err != nil {
		return err
	}
	m.CallTrees[profile] = CallTreeSection{
		Path:        treeRel,
		Purpose:     PurposeCallerCalleeContext,
		Description: "go tool pprof -tree output: caller/callee context for ranked nodes.",
		Producer:    "go tool pprof -tree",
		HotPath:     hotPathSummary(snap.ProfileData, hotPathLimit),
	}
	m.Status.CallTrees[profile] = statusOK

	srcDir := in.Layout.SourceLinesDir(profile, in.Benchmark)
	srcRel, err := in.Layout.RelFromLayout(srcDir)
	if err != nil {
		return err
	}
	m.SourceLines[profile] = SourceLinesSection{
		Dir:         srcRel,
		PathPattern: "source_lines/{profile}/{benchmark}/{output_stem}.txt",
		Purpose:     PurposeLineLevelSource,
		Description: "Per-function go tool pprof -list output with annotated source lines.",
		Producer:    "go tool pprof -list",
		Functions:   functionRefs(in, profile, snap),
	}
	m.Status.SourceLines[profile] = SourceLinesStatus{
		Collected: snap.SourceLinesCollected,
		Skipped:   snap.SourceLinesSkipped,
		Failed:    snap.SourceLinesSkipped,
	}

	pngAbs := in.Layout.CallGraph(profile, in.Benchmark)
	pngRel, err := in.Layout.RelFromLayout(pngAbs)
	if err != nil {
		return err
	}
	ref := CallGraphRef{
		Path:        pngRel,
		Purpose:     PurposeVisualCallGraph,
		Description: "Optional Graphviz PNG from pprof; best-effort during collect.",
	}
	if _, statErr := os.Stat(pngAbs); statErr == nil {
		ref.Status = statusOK
		m.Status.CallGraphs[profile] = CallGraphStatus{Status: statusOK}
	} else {
		ref.Status = statusSkipped
		ref.Reason = "not_generated"
		m.Status.CallGraphs[profile] = CallGraphStatus{Status: statusSkipped, Reason: "not_generated"}
	}
	m.CallGraphs[profile] = ref

	return nil
}

func snapTotal(d *parser.ProfileData) int64 {
	if d == nil {
		return 0
	}
	return d.Total
}

func topSymbols(d *parser.ProfileData, limit int) []TopSymbol {
	if d == nil || len(d.SortedEntries) == 0 {
		return nil
	}
	n := limit
	if len(d.SortedEntries) < n {
		n = len(d.SortedEntries)
	}
	out := make([]TopSymbol, 0, n)
	for i := range n {
		entry := d.SortedEntries[i]
		out = append(out, TopSymbol{
			Rank:    i + 1,
			Symbol:  entry.Name,
			Flat:    entry.Flat,
			Cum:     d.Cum[entry.Name],
			FlatPct: d.FlatPercentages[entry.Name],
			CumPct:  d.CumPercentages[entry.Name],
		})
	}
	return out
}

func hotPathSummary(d *parser.ProfileData, limit int) []string {
	if d == nil {
		return nil
	}
	n := limit
	if len(d.SortedEntries) < n {
		n = len(d.SortedEntries)
	}
	out := make([]string, 0, n)
	for i := range n {
		out = append(out, d.SortedEntries[i].Name)
	}
	return out
}

func functionRefs(in BuildInput, profile string, snap ProfileSnapshot) map[string]FunctionRef {
	functions := make(map[string]FunctionRef, len(snap.ListEntries))
	for _, e := range snap.ListEntries {
		filePath := filepath.Join(in.Layout.SourceLinesDir(profile, in.Benchmark), e.OutputStem+"."+workspace.TextExtension)
		rel, err := in.Layout.RelFromLayout(filePath)
		if err != nil {
			continue
		}
		status := statusOK
		if snap.FailedStems != nil {
			if _, failed := snap.FailedStems[e.OutputStem]; failed {
				status = statusSkipped
			}
		}
		ref := FunctionRef{
			Path:       rel,
			FullSymbol: e.FullSymbol,
			Status:     status,
		}
		if snap.ProfileData != nil {
			ref.Flat = snap.ProfileData.Flat[e.FullSymbol]
			ref.Cum = snap.ProfileData.Cum[e.FullSymbol]
			ref.FlatPct = snap.ProfileData.FlatPercentages[e.FullSymbol]
			ref.CumPct = snap.ProfileData.CumPercentages[e.FullSymbol]
		}
		functions[e.OutputStem] = ref
	}
	if len(functions) == 0 {
		return functions
	}
	// Stable key order not required for map; consumers iterate values.
	return functions
}

// WriteJSON encodes m to path with standard permissions.
func WriteJSON(path string, m BenchmarkMap) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal benchmark map: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(path), workspace.PermDir); err != nil {
		return fmt.Errorf("mkdir map parent: %w", err)
	}
	if err = os.WriteFile(path, data, workspace.PermFile); err != nil {
		return fmt.Errorf("write benchmark map: %w", err)
	}
	return nil
}

// SortedProfileNames returns profile IDs in stable sorted order for tests.
func SortedProfileNames(m BenchmarkMap) []string {
	names := make([]string, 0, len(m.Profiles))
	for name := range m.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

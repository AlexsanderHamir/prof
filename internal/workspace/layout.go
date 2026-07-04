package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// TagLayout is the canonical bench/<tag>/ artifact path contract.
type TagLayout struct {
	Tag  string
	Root string // absolute bench/<tag>/
}

// NewTagLayout builds layout paths under moduleRoot/bench/tag.
func NewTagLayout(moduleRoot, tag string) TagLayout {
	return TagLayout{
		Tag:  tag,
		Root: filepath.Join(moduleRoot, MainDirOutput, tag),
	}
}

// TagLayoutFromCWD resolves module root from cwd and returns the tag layout.
func TagLayoutFromCWD(tag string) (TagLayout, error) {
	root, err := FindModuleRoot()
	if err != nil {
		return TagLayout{}, err
	}
	return NewTagLayout(root, tag), nil
}

// TagDirFromCWD returns bench/<tag>/ under the current module root.
func TagDirFromCWD(tag string) (string, error) {
	l, err := TagLayoutFromCWD(tag)
	if err != nil {
		return "", err
	}
	return l.Root, nil
}

// ProfileBinary returns the raw pprof profile path for a benchmark and profile kind.
func (l TagLayout) ProfileBinary(bench, profile string) string {
	return filepath.Join(l.Root, ProfilesDir, bench, fmt.Sprintf("%s.%s", profile, ProfileArtifactExtension))
}

// Hotspot returns the function-ranked stack summary path for a benchmark and profile kind.
func (l TagLayout) Hotspot(bench, profile string) string {
	return filepath.Join(l.Root, HotspotsDir, bench, fmt.Sprintf("%s.%s", profile, TextExtension))
}

// Measurement returns the go test benchmark run transcript path.
func (l TagLayout) Measurement(bench string) string {
	return filepath.Join(l.Root, MeasurementsDir, bench, MeasurementRunFile)
}

// SourceLinesDir returns the per-function pprof -list output directory.
func (l TagLayout) SourceLinesDir(profile, bench string) string {
	return filepath.Join(l.Root, SourceLinesDir, profile, bench)
}

// CallGraph returns the Graphviz call-graph PNG path for a profile.
func (l TagLayout) CallGraph(profile, bench string) string {
	return filepath.Join(l.Root, CallGraphsDir, profile, bench, fmt.Sprintf("%s.png", profile))
}

// ResolveProfileBinary returns the binary profile path when it exists and is readable.
func (l TagLayout) ResolveProfileBinary(bench, profile string) (string, error) {
	p := l.ProfileBinary(bench, profile)
	if _, err := os.Stat(p); err != nil {
		return "", err
	}
	return p, nil
}

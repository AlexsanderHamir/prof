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

func (l TagLayout) Bin(bench, profile string) string {
	return filepath.Join(l.Root, ProfileBinDir, bench, fmt.Sprintf("%s_%s.%s", bench, profile, ProfileArtifactExtension))
}

func (l TagLayout) Text(bench, profile string) string {
	return filepath.Join(l.Root, ProfileTextDir, bench, fmt.Sprintf("%s_%s.%s", bench, profile, TextExtension))
}

func (l TagLayout) Grouped(bench, profile string) string {
	return filepath.Join(l.Root, ProfileTextDir, bench, fmt.Sprintf("%s_%s_grouped.%s", bench, profile, TextExtension))
}

func (l TagLayout) BenchText(bench string) string {
	return filepath.Join(l.Root, ProfileTextDir, bench, fmt.Sprintf("%s.%s", bench, TextExtension))
}

func (l TagLayout) FunctionsDir(profile, bench string) string {
	return filepath.Join(l.Root, profile+FunctionsDirSuffix, bench)
}

func (l TagLayout) FunctionFile(profile, bench, fnStem string) string {
	return filepath.Join(l.FunctionsDir(profile, bench), fnStem+"."+TextExtension)
}

func (l TagLayout) PNG(profile, bench string) string {
	return filepath.Join(l.FunctionsDir(profile, bench), fmt.Sprintf("%s_%s.png", bench, profile))
}

// ResolveBin returns the binary profile path when it exists and is readable.
func (l TagLayout) ResolveBin(bench, profile string) (string, error) {
	p := l.Bin(bench, profile)
	if _, err := os.Stat(p); err != nil {
		return "", err
	}
	return p, nil
}

// ToolResultsDir returns bench/tools/<toolName>/.
func ToolResultsDir(toolName string) string {
	return filepath.Join(MainDirOutput, ToolDir, toolName)
}

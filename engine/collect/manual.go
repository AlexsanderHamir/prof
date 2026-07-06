package collect

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

// RunManual organizes manual profile files under .prof/<tag>/ using the auto layout.
func RunManual(runner tooling.Runner, opts ManualOptions) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	if err := ensureDirExists(workspace.MainDirOutput); err != nil {
		return err
	}

	tagDir, err := workspace.TagDirFromCWD(opts.Tag)
	if err != nil {
		return err
	}
	if cleanErr := workspace.CleanOrCreateTag(tagDir); cleanErr != nil {
		return fmt.Errorf("CleanOrCreateTag failed: %w", cleanErr)
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	layout, err := workspace.TagLayoutFromCWD(opts.Tag)
	if err != nil {
		return err
	}

	for _, fullBinaryPath := range opts.Files {
		if err = processOneManualFile(runner, fullBinaryPath, layout, cfg); err != nil {
			return err
		}
	}
	return nil
}

func processOneManualFile(runner tooling.Runner, fullBinaryPath string, layout workspace.TagLayout, cfg *config.Config) error {
	benchName, profile := manualBenchAndProfile(fullBinaryPath)
	stem := stemFromPath(fullBinaryPath)
	filter := config.ResolveCollectionFilter(cfg, config.CollectionTargetManual(stem))

	binDest := layout.ProfileBinary(benchName, profile)
	if err := copyProfileBinary(fullBinaryPath, binDest); err != nil {
		return err
	}

	if err := emitProfileArtifacts(runner, binDest, layout, benchName, profile); err != nil {
		return err
	}
	return collectPerFunctionLists(runner, layout, benchName, profile, binDest, filter)
}

func emitProfileArtifacts(runner tooling.Runner, binPath string, layout workspace.TagLayout, benchName, profile string) error {
	hotspotOut := layout.Hotspot(benchName, profile)
	if err := os.MkdirAll(filepath.Dir(hotspotOut), workspace.PermDir); err != nil {
		return fmt.Errorf("mkdir hotspots dest: %w", err)
	}
	return getProfileTextOutput(runner, binPath, hotspotOut)
}

func collectPerFunctionLists(runner tooling.Runner, layout workspace.TagLayout, benchName, profile, binPath string, functionFilter config.FunctionFilter) error {
	listEntries, err := parser.GetFunctionListEntriesV2(binPath, functionFilter)
	if err != nil {
		return fmt.Errorf("extract function names: %w", err)
	}

	functionDir := layout.SourceLinesDir(profile, benchName)
	if err = ensureDirExists(functionDir); err != nil {
		return err
	}
	if err = getFunctionsOutput(runner, listEntries, binPath, functionDir, nil); err != nil {
		return fmt.Errorf("per-function pprof: %w", err)
	}
	return nil
}

func copyProfileBinary(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), workspace.PermDir); err != nil {
		return fmt.Errorf("mkdir profile dest: %w", err)
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open manual profile: %w", err)
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create profile dest: %w", err)
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy profile binary: %w", err)
	}
	return nil
}

func manualBenchAndProfile(fullPath string) (benchName, profile string) {
	stem := stemFromPath(fullPath)
	profile = stem
	benchName = stem
	for _, id := range profileCatalog.ProfileIDsSorted() {
		suffix := "_" + id
		if len(stem) > len(suffix) && stem[len(stem)-len(suffix):] == suffix {
			benchName = stem[:len(stem)-len(suffix)]
			profile = id
			return benchName, profile
		}
	}
	return benchName, profile
}

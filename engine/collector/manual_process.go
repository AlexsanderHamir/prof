package collector

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

// RunCollector organizes manual profile files under bench/<tag>/.
func RunCollector(runner tooling.Runner, files []string, tag string, groupByPackage bool) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	if err := ensureDirExists(internal.MainDirOutput); err != nil {
		return err
	}

	tagDir := filepath.Join(internal.MainDirOutput, tag)
	if err := internal.CleanOrCreateTag(tagDir); err != nil {
		return fmt.Errorf("CleanOrCreateTag failed: %w", err)
	}

	cfg, err := internal.LoadFromFile(internal.ConfigFilename)
	if err != nil {
		cfg = &internal.Config{}
	}

	globalFilter, _ := globalFilterFromConfig(cfg)

	for _, fullBinaryPath := range files {
		if err = processOneManualFile(runner, fullBinaryPath, tagDir, cfg, globalFilter, groupByPackage); err != nil {
			return err
		}
	}
	return nil
}

func processOneManualFile(runner tooling.Runner, fullBinaryPath, tagDir string, cfg *internal.Config, globalFilter internal.FunctionFilter, groupByPackage bool) error {
	name := stemFromPath(fullBinaryPath)
	profileDir, err := profileSubdir(tagDir, name)
	if err != nil {
		return fmt.Errorf("profile directory: %w", err)
	}

	filter := resolveFunctionFilter(cfg, name, globalFilter)
	if err = emitProfileArtifacts(runner, fullBinaryPath, profileDir, name, filter, groupByPackage); err != nil {
		return err
	}
	return collectPerFunctionLists(runner, profileDir, fullBinaryPath, filter)
}

func emitProfileArtifacts(runner tooling.Runner, fullBinaryPath, profileDir, fileName string, filter internal.FunctionFilter, groupByPackage bool) error {
	textOut := filepath.Join(profileDir, fileName+"."+internal.TextExtension)
	if err := GetProfileTextOutput(runner, fullBinaryPath, textOut); err != nil {
		return err
	}
	if groupByPackage {
		grouped := filepath.Join(profileDir, fileName+"_grouped."+internal.TextExtension)
		if err := WriteGroupedPackageProfile(fullBinaryPath, grouped, filter); err != nil {
			return fmt.Errorf("grouped profile: %w", err)
		}
	}
	return nil
}

func collectPerFunctionLists(runner tooling.Runner, profileDirPath, fullBinaryPath string, functionFilter internal.FunctionFilter) error {
	listEntries, err := parser.GetFunctionListEntriesV2(fullBinaryPath, functionFilter)
	if err != nil {
		return fmt.Errorf("extract function names: %w", err)
	}

	functionDir := filepath.Join(profileDirPath, "functions")
	if err = ensureDirExists(functionDir); err != nil {
		return err
	}
	if err = GetFunctionsOutput(runner, listEntries, fullBinaryPath, functionDir); err != nil {
		return fmt.Errorf("per-function pprof: %w", err)
	}
	return nil
}

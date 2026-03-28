package collector

import (
	"fmt"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

// RunCollector organizes manual profile files under bench/<tag>/.
func RunCollector(files []string, tag string, groupByPackage bool) error {
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
		if err = processOneManualFile(fullBinaryPath, tagDir, cfg, globalFilter, groupByPackage); err != nil {
			return err
		}
	}
	return nil
}

func processOneManualFile(fullBinaryPath, tagDir string, cfg *internal.Config, globalFilter internal.FunctionFilter, groupByPackage bool) error {
	name := stemFromPath(fullBinaryPath)
	profileDir, err := profileSubdir(tagDir, name)
	if err != nil {
		return fmt.Errorf("profile directory: %w", err)
	}

	filter := resolveFunctionFilter(cfg, name, globalFilter)
	if err = emitProfileArtifacts(fullBinaryPath, profileDir, name, filter, groupByPackage); err != nil {
		return err
	}
	return collectPerFunctionLists(profileDir, fullBinaryPath, filter)
}

func emitProfileArtifacts(fullBinaryPath, profileDir, fileName string, filter internal.FunctionFilter, groupByPackage bool) error {
	textOut := filepath.Join(profileDir, fileName+"."+internal.TextExtension)
	if err := GetProfileTextOutput(fullBinaryPath, textOut); err != nil {
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

func collectPerFunctionLists(profileDirPath, fullBinaryPath string, functionFilter internal.FunctionFilter) error {
	functions, err := parser.GetAllFunctionNamesV2(fullBinaryPath, functionFilter)
	if err != nil {
		return fmt.Errorf("extract function names: %w", err)
	}

	functionDir := filepath.Join(profileDirPath, "functions")
	if err = ensureDirExists(functionDir); err != nil {
		return err
	}
	if err = GetFunctionsOutput(functions, fullBinaryPath, functionDir); err != nil {
		return fmt.Errorf("per-function pprof: %w", err)
	}
	return nil
}

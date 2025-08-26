package collector

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func ensureDirExists(basePath string) error {
	_, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(basePath, internal.PermDir)
		}
		return err
	}

	return nil
}

// getFunctionPprofContent gets code line level mapping of specified function
// and writes the data to a file named after the function.
func getFunctionPprofContent(function, binaryFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", fmt.Sprintf("-list=%s", function), binaryFile}

	// #nosec ProfileTextDir04 -- cmd is constructed internally by getFunctionPprofContent(), not from user input
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof list command failed: %w", err)
	}

	if err = os.WriteFile(outputFile, output, internal.PermFile); err != nil {
		return fmt.Errorf("failed to write function content: %w", err)
	}

	slog.Info("Collected function", "function", function)
	return nil
}

func collectFunctions(profileDirPath, fullBinaryPath string, functionFilter internal.FunctionFilter) error {
	var functions []string
	functions, err := parser.GetAllFunctionNamesV2(fullBinaryPath, functionFilter)
	if err != nil {
		return fmt.Errorf("failed to extract function names: %w", err)
	}

	functionDir := path.Join(profileDirPath, "functions")
	if err = ensureDirExists(functionDir); err != nil {
		return err
	}

	if err = GetFunctionsOutput(functions, fullBinaryPath, functionDir); err != nil {
		return fmt.Errorf("getAllFunctionsPprofContents failed: %w", err)
	}

	return nil
}

// getGlobalFunctionFilter extracts the global function filter from config
func getGlobalFunctionFilter(cfg *internal.Config) (internal.FunctionFilter, bool) {
	globalFilter, hasGlobalFilter := cfg.FunctionFilter[internal.GlobalSign]
	return globalFilter, hasGlobalFilter
}

// processBinaryFile handles the processing of a single binary file
func processBinaryFile(fullBinaryPath, tagDir string, cfg *internal.Config, globalFilter internal.FunctionFilter, groupByPackage bool) error {
	fileName := getFileName(fullBinaryPath)

	profileDirPath, createErr := createProfileDirectory(tagDir, fileName)
	if createErr != nil {
		return fmt.Errorf("createProfileDirectory failed: %w", createErr)
	}

	functionFilter := determineFunctionFilter(cfg, fileName, globalFilter)

	if genErr := generateProfileOutputs(fullBinaryPath, profileDirPath, fileName, functionFilter, groupByPackage); genErr != nil {
		return genErr
	}

	if collectErr := collectFunctions(profileDirPath, fullBinaryPath, functionFilter); collectErr != nil {
		return fmt.Errorf("collectFunctions failed: %w", collectErr)
	}

	return nil
}

// determineFunctionFilter determines which function filter to use for a given file
func determineFunctionFilter(cfg *internal.Config, fileName string, globalFilter internal.FunctionFilter) internal.FunctionFilter {
	_, hasGlobalFilter := cfg.FunctionFilter[internal.GlobalSign]
	if hasGlobalFilter {
		return globalFilter
	}

	localFilter, hasLocalFilter := cfg.FunctionFilter[fileName]
	if hasLocalFilter {
		return localFilter
	}

	return internal.FunctionFilter{}
}

// generateProfileOutputs generates all profile outputs for a binary file
func generateProfileOutputs(fullBinaryPath, profileDirPath, fileName string, functionFilter internal.FunctionFilter, groupByPackage bool) error {
	outputTextFilePath := path.Join(profileDirPath, fileName+"."+internal.TextExtension)
	if err := GetProfileTextOutput(fullBinaryPath, outputTextFilePath); err != nil {
		return err
	}

	if groupByPackage {
		groupedOutputPath := path.Join(profileDirPath, fileName+"_grouped."+internal.TextExtension)
		if err := generateGroupedProfileData(fullBinaryPath, groupedOutputPath, functionFilter); err != nil {
			return fmt.Errorf("generateGroupedProfileData failed: %w", err)
		}
	}

	return nil
}

// generateGroupedProfileData generates profile data organized by package/module using the new parser function
func generateGroupedProfileData(binaryFile, outputFile string, functionFilter internal.FunctionFilter) error {
	// Import the parser package to use OrganizeProfileByPackageV2
	groupedData, err := parser.OrganizeProfileByPackageV2(binaryFile, functionFilter)
	if err != nil {
		return fmt.Errorf("failed to organize profile by package: %w", err)
	}

	// Write the grouped data to the output file
	return os.WriteFile(outputFile, []byte(groupedData), internal.PermFile)
}

package qcachegrind

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal"
)

// go tool pprof -callgrind {benchmarkName}_{profileName}.out > {benchmarkName}_{profileName}.callgrind
// qcachegrind profile.callgrind

func RunQcacheGrind(tag, benchName, profileName string) error {
	// 1. find the binary file given the parameters of the function, it will be located under bench/tag/bin/benchName/{benchmarkName}_{profileName}.out
	binaryFilePath := filepath.Join(internal.MainDirOutput, tag, internal.ProfileBinDir, benchName, fmt.Sprintf("%s_%s.out", benchName, profileName))

	if _, err := os.Stat(binaryFilePath); os.IsNotExist(err) {
		return fmt.Errorf("binary file not found: %s", binaryFilePath)
	}

	// 2. Create a callgrind file out of the binary by running the following command (EXAMPLE):
	// 	   a. go tool pprof -callgrind {benchmarkName}_{profileName}.out > {benchmarkName}_{profileName}.callgrind
	//     b. save the command under tools/qcachegrind/{benchmarkName}_results.callgrind

	// Create the output directory for qcachegrind results
	outputDir := filepath.Join(internal.MainDirOutput, internal.ToolDir, "qcachegrind")
	if err := os.MkdirAll(outputDir, internal.PermDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the callgrind file
	callgrindOutputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.callgrind", benchName, profileName))

	cmd := exec.Command("go", "tool", "pprof", "-callgrind", binaryFilePath)
	outputFile, err := os.Create(callgrindOutputPath)
	if err != nil {
		return fmt.Errorf("failed to create callgrind output file: %w", err)
	}
	defer outputFile.Close()

	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate callgrind file: %w", err)
	}

	fmt.Printf("Generated callgrind file: %s\n", callgrindOutputPath)

	// 3. Use the output to call qcachegrind profile.callgrind and launch it for the user to analyze.

	// Check if qcachegrind is installed
	if _, err := exec.LookPath("qcachegrind"); err != nil {
		return fmt.Errorf("qcachegrind command not found. Please install it first: sudo apt-get install qcachegrind (Ubuntu/Debian) or brew install qcachegrind (macOS)")
	}

	// Launch qcachegrind with the generated callgrind file
	launchCmd := exec.Command("qcachegrind", callgrindOutputPath)
	launchCmd.Stdout = os.Stdout
	launchCmd.Stderr = os.Stderr

	fmt.Printf("Launching qcachegrind with file: %s\n", callgrindOutputPath)

	// Run qcachegrind in the background so the user can interact with it
	if err := launchCmd.Start(); err != nil {
		return fmt.Errorf("failed to launch qcachegrind: %w", err)
	}

	fmt.Println("qcachegrind launched successfully. You can now analyze the profile data.")

	return nil
}

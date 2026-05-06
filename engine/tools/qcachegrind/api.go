package qcachegrind

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal"
)

// RunQcacheGrind generates callgrind output from a binary profile via go tool pprof and launches qcachegrind.
func RunQcacheGrind(runner tooling.Runner, tag, benchName, profileName string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	// 1. find the binary file given the parameters of the function, it will be located under bench/tag/bin/benchName/{benchmarkName}_{profileName}.out
	binaryFilePath := filepath.Join(internal.MainDirOutput, tag, internal.ProfileBinDir, benchName, fmt.Sprintf("%s_%s.%s", benchName, profileName, internal.ProfileArtifactExtension))

	if _, err := os.Stat(binaryFilePath); os.IsNotExist(err) {
		return fmt.Errorf("binary file not found: %s", binaryFilePath)
	}

	// Create the output directory for qcachegrind results
	outputDir := filepath.Join(internal.MainDirOutput, internal.ToolDir, internal.ToolNameQcachegrind)
	if err := os.MkdirAll(outputDir, internal.PermDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the callgrind file
	callgrindOutputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.callgrind", benchName, profileName))

	outputFile, err := os.Create(callgrindOutputPath)
	if err != nil {
		return fmt.Errorf("failed to create callgrind output file: %w", err)
	}
	defer outputFile.Close()

	ctx := context.Background()
	_, err = runner.Run(ctx, tooling.PprofCallgrindArgs(binaryFilePath), tooling.RunOpts{Stdout: outputFile, Stderr: os.Stderr})
	if err != nil {
		return fmt.Errorf("failed to generate callgrind file: %w", err)
	}

	fmt.Printf("Generated callgrind file: %s\n", callgrindOutputPath)

	if _, err = tooling.LookPath(internal.ToolNameQcachegrind); err != nil {
		return errors.New("qcachegrind command not found. Please install it first: sudo apt-get install qcachegrind (Ubuntu/Debian) or brew install qcachegrind (macOS)")
	}

	fmt.Printf("Launching qcachegrind with file: %s\n", callgrindOutputPath)

	if err = tooling.StartDetached(ctx, []string{internal.ToolNameQcachegrind, callgrindOutputPath}, tooling.RunOpts{}); err != nil {
		return fmt.Errorf("failed to launch qcachegrind: %w", err)
	}

	fmt.Println("qcachegrind launched successfully. You can now analyze the profile data.")

	return nil
}

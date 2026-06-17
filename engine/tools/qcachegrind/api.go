package qcachegrind

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// RunQcacheGrind generates callgrind output from a binary profile via go tool pprof and launches qcachegrind.
func RunQcacheGrind(runner tooling.Runner, tag, benchName, profileName string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	layout := workspace.NewTagLayout(cwd, tag)
	binaryFilePath := layout.Bin(benchName, profileName)

	if _, statErr := os.Stat(binaryFilePath); os.IsNotExist(statErr) {
		return fmt.Errorf("binary file not found: %s", binaryFilePath)
	}

	outputDir := filepath.Join(workspace.MainDirOutput, workspace.ToolDir, workspace.ToolNameQcachegrind)
	if mkdirErr := os.MkdirAll(outputDir, workspace.PermDir); mkdirErr != nil {
		return fmt.Errorf("failed to create output directory: %w", mkdirErr)
	}

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

	if _, err = tooling.LookPath(workspace.ToolNameQcachegrind); err != nil {
		return errors.New("qcachegrind command not found. Please install it first: sudo apt-get install qcachegrind (Ubuntu/Debian) or brew install qcachegrind (macOS)")
	}

	fmt.Printf("Launching qcachegrind with file: %s\n", callgrindOutputPath)

	if err = tooling.StartDetached(ctx, []string{workspace.ToolNameQcachegrind, callgrindOutputPath}, tooling.RunOpts{}); err != nil {
		return fmt.Errorf("failed to launch qcachegrind: %w", err)
	}

	fmt.Println("qcachegrind launched successfully. You can now analyze the profile data.")
	return nil
}

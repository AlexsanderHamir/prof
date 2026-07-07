package collect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

func runPprofReport(runner tooling.Runner, argv []string, outputFile string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	ctx := context.Background()
	out, err := runner.Run(ctx, argv, tooling.RunOpts{})
	if err != nil {
		return fmt.Errorf("pprof command failed: %w", err)
	}
	return writeArtifactFile(outputFile, out)
}

func getPNGOutput(runner tooling.Runner, binaryFile, outputFile string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	ctx := context.Background()
	out, err := runner.Run(ctx, tooling.PprofPNGArgs(binaryFile), tooling.RunOpts{})
	if err != nil {
		return fmt.Errorf("pprof PNG generation failed: %w", err)
	}
	return writeArtifactFile(outputFile, out)
}

func writeArtifactFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), workspace.PermDir); err != nil {
		return fmt.Errorf("mkdir artifact parent: %w", err)
	}
	return os.WriteFile(path, data, workspace.PermFile)
}

func listPatternCandidates(shortStem, fullSymbol string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if fullSymbol != "" {
		add(regexp.QuoteMeta(fullSymbol))
	}
	add(shortStem)
	return out
}

func writeFunctionListPprof(runner tooling.Runner, shortStem, fullSymbol, binaryFile, outputFile string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	ctx := context.Background()
	var lastErr error
	for _, pattern := range listPatternCandidates(shortStem, fullSymbol) {
		out, err := runner.Run(ctx, tooling.PprofListArgs(binaryFile, pattern), tooling.RunOpts{Combined: true})
		if err != nil {
			lastErr = fmt.Errorf("pprof list (pattern %q): %w: %s", pattern, err, string(out))
			continue
		}
		if err = os.WriteFile(outputFile, out, workspace.PermFile); err != nil {
			return fmt.Errorf("write function content: %w", err)
		}
		return nil
	}
	return lastErr
}

// ListResult summarizes per-function pprof -list collection for one profile.
type ListResult struct {
	Collected   int
	Skipped     int
	FailedStems map[string]struct{}
}

func getFunctionsOutput(runner tooling.Runner, entries []parser.FunctionListEntry, binaryPath, basePath string, session *termui.Session) ListResult {
	const maxPerFunctionWarnings = 3

	result := ListResult{FailedStems: make(map[string]struct{})}
	errs := parallelFor(len(entries), sourceLinesWorkers(len(entries)), func(i int) error {
		e := entries[i]
		out := filepath.Join(basePath, e.OutputStem+"."+workspace.TextExtension)
		return writeFunctionListPprof(runner, e.OutputStem, e.FullSymbol, binaryPath, out)
	})

	for i, err := range errs {
		if err == nil {
			result.Collected++
			continue
		}
		result.Skipped++
		result.FailedStems[entries[i].OutputStem] = struct{}{}
		if session != nil && session.Interactive() && result.Skipped <= maxPerFunctionWarnings {
			session.Warn(fmt.Sprintf("skipping per-function pprof list for %s: %v", entries[i].OutputStem, err))
		}
	}
	if result.Skipped > maxPerFunctionWarnings && session != nil && session.Interactive() {
		session.Warn(fmt.Sprintf("… and %d more functions skipped", result.Skipped-maxPerFunctionWarnings))
	}
	return result
}

// FunctionsOutput runs pprof -list for each entry (exported for integration tests).
func FunctionsOutput(runner tooling.Runner, entries []parser.FunctionListEntry, binaryPath, basePath string) error {
	_ = getFunctionsOutput(runner, entries, binaryPath, basePath, nil)
	return nil
}

package collect

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

func getProfileTextOutput(runner tooling.Runner, binaryFile, outputFile string) error {
	if runner == nil {
		return errors.New("tooling runner is nil")
	}
	ctx := context.Background()
	out, err := runner.Run(ctx, tooling.PprofTextTopArgs(binaryFile), tooling.RunOpts{})
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
		slog.Info("Collected function", "function", shortStem, "list_pattern", pattern)
		return nil
	}
	return lastErr
}

func getFunctionsOutput(runner tooling.Runner, entries []parser.FunctionListEntry, binaryPath, basePath string) error {
	for _, e := range entries {
		out := filepath.Join(basePath, e.OutputStem+"."+workspace.TextExtension)
		if err := writeFunctionListPprof(runner, e.OutputStem, e.FullSymbol, binaryPath, out); err != nil {
			slog.Warn("skipping per-function pprof list", "function", e.OutputStem, "binary", binaryPath, "err", err)
			continue
		}
	}
	return nil
}

// FunctionsOutput runs pprof -list for each entry (exported for integration tests).
func FunctionsOutput(runner tooling.Runner, entries []parser.FunctionListEntry, binaryPath, basePath string) error {
	return getFunctionsOutput(runner, entries, binaryPath, basePath)
}

func writeGroupedPackageProfile(binaryPath, outputPath string, filter config.FunctionFilter) error {
	text, err := parser.OrganizeProfileByPackageV2(binaryPath, filter)
	if err != nil {
		return fmt.Errorf("organize profile by package: %w", err)
	}
	return writeArtifactFile(outputPath, []byte(text))
}

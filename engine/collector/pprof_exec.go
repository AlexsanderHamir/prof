package collector

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

// GetProfileTextOutput runs go tool pprof text listing and writes output to outputFile.
func GetProfileTextOutput(binaryFile, outputFile string) error {
	cmd := append([]string{"go", "tool", "pprof"}, pprofTextListArgs()...)
	cmd = append(cmd, binaryFile)
	// #nosec G204 -- argv built here, binary path only
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof command failed: %w", err)
	}
	return os.WriteFile(outputFile, out, internal.PermFile)
}

// GetPNGOutput renders a PNG flame-style view via go tool pprof -png.
func GetPNGOutput(binaryFile, outputFile string) error {
	cmd := []string{"go", "tool", "pprof", "-png", binaryFile}
	// #nosec G204
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return fmt.Errorf("pprof PNG generation failed: %w", err)
	}
	return os.WriteFile(outputFile, out, internal.PermFile)
}

// listPatternCandidates returns -list= regexp arguments to try, most specific first.
// pprof matches the pattern against merged profile graph node names; a literal
// full symbol (QuoteMeta) succeeds where a short basename alone does not.
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

func writeFunctionListPprof(shortStem, fullSymbol, binaryFile, outputFile string) error {
	var lastErr error
	for _, pattern := range listPatternCandidates(shortStem, fullSymbol) {
		cmd := []string{"go", "tool", "pprof", "-list=" + pattern, binaryFile}
		// #nosec G204
		execCmd := exec.Command(cmd[0], cmd[1:]...)
		out, err := execCmd.CombinedOutput()
		if err != nil {
			lastErr = fmt.Errorf("pprof list (pattern %q): %w: %s", pattern, err, string(out))
			continue
		}
		if err = os.WriteFile(outputFile, out, internal.PermFile); err != nil {
			return fmt.Errorf("write function content: %w", err)
		}
		slog.Info("Collected function", "function", shortStem, "list_pattern", pattern)
		return nil
	}
	return lastErr
}

// GetFunctionsOutput runs pprof -list for each [parser.FunctionListEntry] into basePath.
// If every -list pattern fails for a symbol, that function is skipped and a warning is logged
// so the rest of the profile can still be collected.
func GetFunctionsOutput(entries []parser.FunctionListEntry, binaryPath, basePath string) error {
	for _, e := range entries {
		out := filepath.Join(basePath, e.OutputStem+"."+internal.TextExtension)
		if err := writeFunctionListPprof(e.OutputStem, e.FullSymbol, binaryPath, out); err != nil {
			slog.Warn("skipping per-function pprof list", "function", e.OutputStem, "binary", binaryPath, "err", err)
			continue
		}
	}
	return nil
}

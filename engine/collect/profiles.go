package collect

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func processProfiles(runner tooling.Runner, benchmarkName string, profiles []string, tag string, lenientProfiles bool, skipPNG bool, session termui.Session) ([]string, error) { //nolint:gocognit // sequential profile stages
	layout, err := workspace.TagLayoutFromCWD(tag)
	if err != nil {
		return nil, err
	}

	var processed []string

	for _, profile := range profiles {
		profileFile := layout.Bin(benchmarkName, profile)
		if _, statErr := os.Stat(profileFile); statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				if lenientProfiles {
					warnMissingProfile(session, profileFile)
					continue
				}
				return nil, fmt.Errorf("missing profile binary for benchmark %s profile %s: %w", benchmarkName, profile, statErr)
			}
			return nil, fmt.Errorf("failed to stat profile file %s: %w", profileFile, statErr)
		}

		outputFile := layout.Text(benchmarkName, profile)
		fnDir := layout.FunctionsDir(profile, benchmarkName)

		if textErr := getProfileTextOutput(runner, profileFile, outputFile); textErr != nil {
			return nil, fmt.Errorf("failed to generate text profile for %s: %w", profile, textErr)
		}

		if mkdirErr := os.MkdirAll(fnDir, workspace.PermDir); mkdirErr != nil {
			return nil, fmt.Errorf("failed to create profile functions directory: %w", mkdirErr)
		}

		pngPath := layout.PNG(profile, benchmarkName)
		if pngErr := getPNGOutput(runner, profileFile, pngPath); pngErr != nil {
			if skipPNG {
				warnSkippedPNG(session, profile, benchmarkName, pngErr)
			} else {
				return nil, fmt.Errorf("failed to generate PNG for profile %s (install graphviz or use --skip-png): %w", profile, pngErr)
			}
		}

		if !session.Interactive() {
			slog.Info("Processed profile", "profile", profile, "benchmark", benchmarkName)
		}
		processed = append(processed, profile)
	}

	if len(processed) == 0 && len(profiles) > 0 {
		return nil, fmt.Errorf("no profile binaries processed for benchmark %s", benchmarkName)
	}

	return processed, nil
}

func warnMissingProfile(session termui.Session, profileFile string) {
	msg := fmt.Sprintf("profile file not found, skipping: %s", profileFile)
	if session.Interactive() {
		session.Warn(msg)
		return
	}
	slog.Warn("Profile file not found — skipping", "file", profileFile)
}

func warnSkippedPNG(session termui.Session, profile, benchmarkName string, pngErr error) {
	msg := fmt.Sprintf("PNG skipped for %s/%s: %v", benchmarkName, profile, pngErr)
	if session.Interactive() {
		session.Warn(msg)
		return
	}
	slog.Warn("PNG visualization skipped", "profile", profile, "benchmark", benchmarkName, "err", pngErr)
}

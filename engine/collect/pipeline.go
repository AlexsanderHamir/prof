package collect

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

func runBenchAndGetProfiles(runner tooling.Runner, autoArgs *config.AutoArgs, cfg *config.Config, session *termui.Session) error {
	if !session.Interactive() {
		slog.Info("Starting benchmark pipeline...")
	}

	total := len(autoArgs.Benchmarks)
	profileDetail := strings.Join(autoArgs.Profiles, ", ")
	countDetail := fmt.Sprintf("count=%d", autoArgs.Count)

	for i, benchmarkName := range autoArgs.Benchmarks {
		base := termui.Progress{
			Label: benchmarkName,
			Index: i + 1,
			Total: total,
		}

		if !session.Interactive() {
			slog.Info("Running benchmark", "Benchmark", benchmarkName)
		}
		if session.Interactive() {
			session.BeginBenchmark(i+1, total, benchmarkName)
		}
		if err := session.RunWhile(base.WithPhase(termui.PhaseRunBenchmark).WithDetail(countDetail), func() error {
			return runBenchmark(runner, benchmarkName, autoArgs.Profiles, autoArgs.Count, autoArgs.Tag)
		}); err != nil {
			return fmt.Errorf("failed to run %s: %w", benchmarkName, err)
		}

		filter := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto(benchmarkName))

		if !session.Interactive() {
			slog.Info("Processing profiles", "Benchmark", benchmarkName)
		}
		var profilesReady []string
		if err := session.RunWhile(base.WithPhase(termui.PhaseCollectProfiles).WithDetail(profileDetail), func() error {
			var procErr error
			profilesReady, procErr = processProfiles(runner, benchmarkName, autoArgs.Profiles, autoArgs.Tag, session)
			return procErr
		}); err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		if !session.Interactive() {
			slog.Info("Collecting function profiles", "Benchmark", benchmarkName)
		}
		args := &config.CollectionArgs{
			Tag:             autoArgs.Tag,
			Profiles:        profilesReady,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: filter,
		}
		if err := session.RunWhile(base.WithPhase(termui.PhaseCollectFunctionProfiles), func() error {
			return collectProfileFunctions(runner, args, session)
		}); err != nil {
			return fmt.Errorf("failed to collect function profiles for %s: %w", benchmarkName, err)
		}

		if !session.Interactive() {
			slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
		}
	}

	session.Success(workspace.InfoCollectionSuccess)
	return nil
}

func collectProfileFunctions(runner tooling.Runner, args *config.CollectionArgs, session *termui.Session) error {
	layout, err := workspace.TagLayoutFromCWD(args.Tag)
	if err != nil {
		return err
	}

	for _, profile := range args.Profiles {
		fnDir := layout.SourceLinesDir(profile, args.BenchmarkName)
		if mkdirErr := os.MkdirAll(fnDir, workspace.PermDir); mkdirErr != nil {
			return fmt.Errorf("failed to create output directory: %w", mkdirErr)
		}

		binPath := layout.ProfileBinary(args.BenchmarkName, profile)
		listEntries, listErr := parser.GetFunctionListEntriesV2(binPath, args.BenchmarkConfig)
		if listErr != nil {
			return fmt.Errorf("failed to extract function names: %w", listErr)
		}

		if fnErr := getFunctionsOutput(runner, listEntries, binPath, fnDir, session); fnErr != nil {
			return fmt.Errorf("getAllFunctionsPprofContents failed: %w", fnErr)
		}
	}

	return nil
}

package collect

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/datamap"
	"github.com/AlexsanderHamir/prof/internal/termui"
	"github.com/AlexsanderHamir/prof/internal/workspace"
	"github.com/AlexsanderHamir/prof/parser"
)

func finalizeInteractiveErr(session *termui.Session, err error) error {
	if err == nil || session == nil || !session.Interactive() || !session.ErrorDisplayed() {
		return err
	}
	return termui.StagedDisplay(err)
}

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
			return finalizeInteractiveErr(session, fmt.Errorf("failed to run %s: %w", benchmarkName, err))
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
			return finalizeInteractiveErr(session, fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err))
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
			return collectFunctionsAndEmitMap(runner, args, session, autoArgs, benchmarkName, filter, profilesReady)
		}); err != nil {
			return finalizeInteractiveErr(session, fmt.Errorf("failed to collect function profiles for %s: %w", benchmarkName, err))
		}

		if !session.Interactive() {
			slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
		}
	}

	session.Success(workspace.InfoCollectionSuccess)
	return nil
}

func collectFunctionsAndEmitMap(
	runner tooling.Runner,
	args *config.CollectionArgs,
	session *termui.Session,
	autoArgs *config.AutoArgs,
	benchmarkName string,
	filter config.FunctionFilter,
	profilesReady []string,
) error {
	snapshots, err := collectProfileFunctions(runner, args, session)
	if err != nil {
		return err
	}
	layout, err := workspace.TagLayoutFromCWD(autoArgs.Tag)
	if err != nil {
		return err
	}
	emitBenchmarkMap(session, layout, emitMapParams{
		Tag:              autoArgs.Tag,
		Benchmark:        benchmarkName,
		Profiles:         profilesReady,
		Filter:           filter,
		BenchCount:       autoArgs.Count,
		CollectionMode:   datamapCollectionAuto,
		PerProfile:       snapshots,
		IncludeMeasuring: true,
	})
	return nil
}

func collectProfileFunctions(runner tooling.Runner, args *config.CollectionArgs, session *termui.Session) ([]datamap.ProfileSnapshot, error) {
	layout, err := workspace.TagLayoutFromCWD(args.Tag)
	if err != nil {
		return nil, err
	}

	profiles := args.Profiles
	snapshots := make([]datamap.ProfileSnapshot, len(profiles))
	errs := parallelFor(len(profiles), sourceLinesWorkers(len(profiles)), func(i int) error {
		profile := profiles[i]
		fnDir := layout.SourceLinesDir(profile, args.BenchmarkName)
		if mkdirErr := os.MkdirAll(fnDir, workspace.PermDir); mkdirErr != nil {
			return fmt.Errorf("failed to create output directory: %w", mkdirErr)
		}

		binPath := layout.ProfileBinary(args.BenchmarkName, profile)
		listEntries, profileData, listErr := parser.GetFunctionListEntriesWithProfileData(binPath, args.BenchmarkConfig)
		if listErr != nil {
			return fmt.Errorf("failed to extract function names: %w", listErr)
		}

		listResult := getFunctionsOutput(runner, listEntries, binPath, fnDir, session)
		snapshots[i] = datamap.ProfileSnapshot{
			Profile:              profile,
			ProfileData:          profileData,
			ListEntries:          listEntries,
			SourceLinesCollected: listResult.Collected,
			SourceLinesSkipped:   listResult.Skipped,
			FailedStems:          listResult.FailedStems,
		}
		return nil
	})

	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("profile %s: %w", profiles[i], err)
		}
	}
	return snapshots, nil
}

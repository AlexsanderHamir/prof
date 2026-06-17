package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/intent"
)

const (
	cfgOverviewCollection = "Profile collection — what to save from benchmarks"
	cfgOverviewTrack      = "Regression limits — when a compare should fail"
	cfgOverviewViewPath   = "View prof.json file path"
	cfgOverviewSave       = "Save and exit"
	cfgOverviewDiscard    = "Discard changes and exit"
)

func runUIConfigWizard(svc *app.Services) error {
	path, err := svc.Config.Path()
	if err != nil {
		return err
	}

	var loadedAt time.Time
	cfg, err := loadConfigForWizard(svc, path, &loadedAt)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "Settings edit prof.json next to go.mod.")
	fmt.Fprintln(os.Stdout, "Defaults work for most projects — choose Save and exit unless you need advanced filters or CI limits.")

	for {
		choice, menuErr := configOverviewMenu(cfg)
		if menuErr != nil {
			return menuErr
		}
		switch choice {
		case cfgOverviewCollection:
			if err = runCollectionSubmenu(svc, cfg); err != nil {
				return err
			}
		case cfgOverviewTrack:
			if err = runTrackSubmenu(svc, cfg); err != nil {
				return err
			}
		case cfgOverviewViewPath:
			fmt.Fprintf(os.Stdout, "Configuration file:\n  %s\n", path)
		case cfgOverviewSave:
			if err = confirmSaveIfChanged(path, loadedAt); err != nil {
				return err
			}
			config.Normalize(cfg)
			if err = intent.RunValidated(&intent.ConfigSaveIntent{Config: cfg}, svc); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Saved configuration to %s\n", path)
			return nil
		case cfgOverviewDiscard:
			return nil
		default:
			return fmt.Errorf("unknown menu choice: %s", choice)
		}
	}
}

func loadConfigForWizard(svc *app.Services, path string, loadedAt *time.Time) (*config.Config, error) {
	info, err := ensureConfigFileExists(svc, path)
	if err != nil {
		return nil, err
	}
	*loadedAt = info.ModTime()
	return loadValidConfig(svc, path, loadedAt)
}

func ensureConfigFileExists(svc *app.Services, path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	return promptCreateConfigFile(svc, path)
}

func promptCreateConfigFile(svc *app.Services, path string) (os.FileInfo, error) {
	create := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "No prof.json found. Create a default configuration file?",
		Default: true,
	}, &create); err != nil {
		return nil, err
	}
	if !create {
		return nil, errors.New("configuration cancelled: prof.json does not exist")
	}
	if err := intent.RunValidated(&intent.ConfigCreateIntent{}, svc); err != nil {
		return nil, err
	}
	return os.Stat(path)
}

func loadValidConfig(svc *app.Services, path string, loadedAt *time.Time) (*config.Config, error) {
	cfg, err := svc.Config.Load()
	if err == nil {
		return cfg, nil
	}
	return recoverInvalidConfigFile(svc, path, loadedAt, err)
}

func recoverInvalidConfigFile(svc *app.Services, path string, loadedAt *time.Time, loadErr error) (*config.Config, error) {
	fix := false
	fmt.Fprintf(os.Stderr, "Invalid prof.json: %v\n", loadErr)
	if askErr := survey.AskOne(&survey.Confirm{
		Message: "Back up to prof.json.bak and recreate from defaults?",
		Default: false,
	}, &fix); askErr != nil {
		return nil, askErr
	}
	if !fix {
		return nil, loadErr
	}
	if backupErr := os.Rename(path, path+".bak"); backupErr != nil {
		return nil, backupErr
	}
	if createErr := intent.RunValidated(&intent.ConfigCreateIntent{}, svc); createErr != nil {
		return nil, createErr
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		return nil, statErr
	}
	*loadedAt = info.ModTime()
	return svc.Config.Load()
}

func configOverviewMenu(cfg *config.Config) (string, error) {
	nBench := len(cfg.Collection.Benchmarks)
	nManual := len(cfg.Collection.ManualProfiles)
	nTrack := len(cfg.Track.Benchmarks)
	options := []string{
		fmt.Sprintf("%s (defaults + %d benchmarks + %d manual profiles)", cfgOverviewCollection, nBench, nManual),
		fmt.Sprintf("%s (defaults + %d benchmark overrides)", cfgOverviewTrack, nTrack),
		cfgOverviewViewPath,
		cfgOverviewSave,
		cfgOverviewDiscard,
	}
	var choice string
	err := survey.AskOne(&survey.Select{
		Message:  "Settings — what do you want to change?",
		Options:  options,
		PageSize: 8,
		Help:     "Save and exit writes prof.json. You can skip the other sections if defaults are fine.",
	}, &choice, survey.WithValidator(survey.Required))
	if err != nil {
		return "", err
	}
	switch {
	case strings.HasPrefix(choice, cfgOverviewCollection):
		return cfgOverviewCollection, nil
	case strings.HasPrefix(choice, cfgOverviewTrack):
		return cfgOverviewTrack, nil
	default:
		return choice, nil
	}
}

func runCollectionSubmenu(svc *app.Services, cfg *config.Config) error {
	const (
		done         = "Done — defaults are fine"
		editDefaults = "Advanced: change default package and function filters"
		editBench    = "Advanced: filters for one benchmark only"
		editManual   = "Advanced: filters for one manual profile only"
		removeRule   = "Remove a benchmark or manual rule"
		back         = "Back to settings menu"
	)
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Profile collection controls which functions appear in saved files after you run benchmarks.")
	fmt.Fprintln(os.Stdout, "Most projects: choose Done — defaults are fine. Your module path from go.mod is already included.")
	for {
		var choice string
		if err := survey.AskOne(&survey.Select{
			Message:  "Profile collection",
			Options:  []string{done, editDefaults, editBench, editManual, removeRule, back},
			PageSize: 8,
		}, &choice, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		switch choice {
		case done, back:
			return nil
		case editDefaults:
			if err := editFunctionFilter("Collection defaults", &cfg.Collection.Defaults, svc); err != nil {
				return err
			}
		case editBench:
			if err := editCollectionBenchmarkRule(svc, cfg); err != nil {
				return err
			}
		case editManual:
			if err := editCollectionManualRule(svc, cfg); err != nil {
				return err
			}
		case removeRule:
			if err := removeCollectionRule(cfg); err != nil {
				return err
			}
		}
	}
}

func runTrackSubmenu(svc *app.Services, cfg *config.Config) error {
	const (
		done         = "Done — defaults are fine"
		editDefaults = "Advanced: change global regression limits"
		editBench    = "Advanced: limits for one benchmark only"
		removeBench  = "Remove a benchmark override"
		back         = "Back to settings menu"
	)
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Regression limits control when the built-in compare run fails (for example in CI).")
	fmt.Fprintln(os.Stdout, "Most projects: choose Done — defaults are fine, or adjust fail thresholds under Advanced.")
	for {
		var choice string
		if err := survey.AskOne(&survey.Select{
			Message:  "Regression limits",
			Options:  []string{done, editDefaults, editBench, removeBench, back},
			PageSize: 6,
		}, &choice, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		switch choice {
		case done, back:
			return nil
		case editDefaults:
			if err := editTrackPolicy("Track defaults", &cfg.Track.Defaults); err != nil {
				return err
			}
		case editBench:
			if err := editTrackBenchmarkOverride(svc, cfg); err != nil {
				return err
			}
		case removeBench:
			if err := removeTrackBenchmark(cfg); err != nil {
				return err
			}
		}
	}
}

func editCollectionBenchmarkRule(svc *app.Services, cfg *config.Config) error {
	names, err := svc.Collect.DiscoverBenchmarks("")
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return errors.New("no benchmarks found in module (look for func BenchmarkXxx in *_test.go)")
	}
	var bench string
	if err = survey.AskOne(&survey.Select{
		Message:  "Benchmark name:",
		Options:  names,
		PageSize: tuiPageSize,
	}, &bench, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if cfg.Collection.Benchmarks == nil {
		cfg.Collection.Benchmarks = map[string]config.FunctionFilter{}
	}
	f := cfg.Collection.Benchmarks[bench]
	if err = editFunctionFilter("Benchmark "+bench, &f, svc); err != nil {
		return err
	}
	cfg.Collection.Benchmarks[bench] = f
	return nil
}

func editCollectionManualRule(svc *app.Services, cfg *config.Config) error {
	const customKey = "Custom key (type manually)…"
	keys := manualProfileKeyOptions(svc)
	options := make([]string, 0, len(keys)+1)
	options = append(options, keys...)
	options = append(options, customKey)
	var pick string
	if err := survey.AskOne(&survey.Select{
		Message:  "Manual profile key:",
		Options:  options,
		PageSize: tuiPageSize,
	}, &pick, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	key := pick
	if pick == customKey {
		fmt.Fprintln(os.Stdout, "Keys match the file stem, e.g. BenchmarkFoo_cpu for BenchmarkFoo_cpu.out")
		if err := survey.AskOne(&survey.Input{
			Message: "Manual profile key:",
		}, &key, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}
	key = strings.TrimSpace(key)
	if cfg.Collection.ManualProfiles == nil {
		cfg.Collection.ManualProfiles = map[string]config.FunctionFilter{}
	}
	f := cfg.Collection.ManualProfiles[key]
	if err := editFunctionFilter("Manual profile "+key, &f, nil); err != nil {
		return err
	}
	cfg.Collection.ManualProfiles[key] = f
	return nil
}

func manualProfileKeyOptions(svc *app.Services) []string {
	benches, err := svc.Collect.DiscoverBenchmarks("")
	if err != nil || len(benches) == 0 {
		return nil
	}
	profiles := svc.Collect.SupportedProfiles()
	var keys []string
	for _, bench := range benches {
		for _, profile := range profiles {
			keys = append(keys, bench+"_"+profile)
		}
	}
	return keys
}

func removeCollectionRule(cfg *config.Config) error {
	var keys []string
	for k := range cfg.Collection.Benchmarks {
		keys = append(keys, "benchmark: "+k)
	}
	for k := range cfg.Collection.ManualProfiles {
		keys = append(keys, "manual: "+k)
	}
	if len(keys) == 0 {
		fmt.Fprintln(os.Stdout, "No benchmark or manual rules to remove.")
		return nil
	}
	var pick string
	if err := survey.AskOne(&survey.Select{
		Message:  "Remove which rule?",
		Options:  keys,
		PageSize: tuiPageSize,
	}, &pick, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	confirm := false
	if err := survey.AskOne(&survey.Confirm{Message: "Remove this rule?", Default: false}, &confirm); err != nil {
		return err
	}
	if !confirm {
		return nil
	}
	if strings.HasPrefix(pick, "benchmark: ") {
		delete(cfg.Collection.Benchmarks, strings.TrimPrefix(pick, "benchmark: "))
	}
	if strings.HasPrefix(pick, "manual: ") {
		delete(cfg.Collection.ManualProfiles, strings.TrimPrefix(pick, "manual: "))
	}
	return nil
}

func editTrackBenchmarkOverride(svc *app.Services, cfg *config.Config) error {
	names, err := svc.Collect.DiscoverBenchmarks("")
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return errors.New("no benchmarks found in module (look for func BenchmarkXxx in *_test.go)")
	}
	var bench string
	if err = survey.AskOne(&survey.Select{
		Message:  "Benchmark name to override:",
		Options:  names,
		PageSize: tuiPageSize,
	}, &bench, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if cfg.Track.Benchmarks == nil {
		cfg.Track.Benchmarks = map[string]config.TrackPolicy{}
	}
	p := cfg.Track.Benchmarks[bench]
	if editErr := editTrackPolicy("Benchmark override "+bench, &p); editErr != nil {
		return editErr
	}
	cfg.Track.Benchmarks[bench] = p
	return nil
}

func removeTrackBenchmark(cfg *config.Config) error {
	if len(cfg.Track.Benchmarks) == 0 {
		fmt.Fprintln(os.Stdout, "No benchmark overrides to remove.")
		return nil
	}
	keys := make([]string, 0, len(cfg.Track.Benchmarks))
	for k := range cfg.Track.Benchmarks {
		keys = append(keys, k)
	}
	var pick string
	if err := survey.AskOne(&survey.Select{
		Message:  "Remove override for:",
		Options:  keys,
		PageSize: tuiPageSize,
	}, &pick, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	delete(cfg.Track.Benchmarks, pick)
	return nil
}

func editFunctionFilter(label string, f *config.FunctionFilter, svc *app.Services) error {
	defaultPrefix := strings.Join(f.IncludePrefixes, ", ")
	if defaultPrefix == "" && svc != nil {
		if cfg, err := config.DefaultFromModuleRoot(); err == nil && len(cfg.Collection.Defaults.IncludePrefixes) > 0 {
			defaultPrefix = cfg.Collection.Defaults.IncludePrefixes[0]
		}
	}

	fmt.Fprintln(os.Stdout, "")
	if label != "Collection defaults" {
		fmt.Fprintf(os.Stdout, "Editing: %s\n", label)
	}
	fmt.Fprintln(os.Stdout, "Press Enter at each prompt to keep the suggested value.")
	fmt.Fprintln(os.Stdout, "Package filter: limits saved profiles to your code (import paths from go.mod), not the Go standard library.")
	fmt.Fprintln(os.Stdout, "Function filter: optional names to skip, such as init or test helpers.")

	var prefixes string
	if err := survey.AskOne(&survey.Input{
		Message: "Go package import paths to include (comma-separated):",
		Default: defaultPrefix,
		Help:    "Example: github.com/you/yourapp. Leave empty only if you want every function in the profile, including stdlib.",
	}, &prefixes); err != nil {
		return err
	}
	f.IncludePrefixes = splitCSV(prefixes)

	var ignores string
	if err := survey.AskOne(&survey.Input{
		Message: "Function names to skip (comma-separated, optional):",
		Default: strings.Join(f.IgnoreFunctions, ", "),
		Help:    "Example: init, BenchmarkHelper. Empty means do not skip any names beyond the package filter.",
	}, &ignores); err != nil {
		return err
	}
	f.IgnoreFunctions = splitCSV(ignores)
	return nil
}

func editTrackPolicy(label string, p *config.TrackPolicy) error {
	fmt.Fprintln(os.Stdout, "")
	if label != "Track defaults" {
		fmt.Fprintf(os.Stdout, "Editing: %s\n", label)
	}
	fmt.Fprintln(os.Stdout, "These apply when you use the built-in regression check. Press Enter to keep each suggested value.")
	fmt.Fprintln(os.Stdout, "CLI flags can override these for a single run.")

	var prefixes string
	if err := survey.AskOne(&survey.Input{
		Message: "Skip functions whose names start with (comma-separated, optional):",
		Default: strings.Join(p.IgnorePrefixes, ", "),
		Help:    "Example: runtime., testing. — common noise from the Go runtime and test framework.",
	}, &prefixes); err != nil {
		return err
	}
	p.IgnorePrefixes = splitCSV(prefixes)

	var funcs string
	if err := survey.AskOne(&survey.Input{
		Message: "Skip exact function names (comma-separated, optional):",
		Default: strings.Join(p.IgnoreFunctions, ", "),
		Help:    "Example: init. Empty means only use the prefix list above.",
	}, &funcs); err != nil {
		return err
	}
	p.IgnoreFunctions = splitCSV(funcs)

	var minStr string
	if err := survey.AskOne(&survey.Input{
		Message: "Ignore slowdowns smaller than this percent (0 = report everything):",
		Default: fmt.Sprintf("%.1f", p.MinChangePercent),
		Help:    "Example: 5 means a 3% slowdown is treated as noise and ignored.",
	}, &minStr); err != nil {
		return err
	}
	minPct, parseErr := strconv.ParseFloat(strings.TrimSpace(minStr), 64)
	if parseErr != nil || minPct < 0 {
		return fmt.Errorf("invalid min_change_percent: %s", minStr)
	}
	p.MinChangePercent = minPct

	var maxStr string
	if askErr := survey.AskOne(&survey.Input{
		Message: "Fail the run if slowdown exceeds this percent (0 = never fail on slowdown):",
		Default: fmt.Sprintf("%.1f", p.MaxRegressionPercent),
		Help:    "Example: 15 fails CI when a function is 15% or more slower than the baseline run.",
	}, &maxStr); askErr != nil {
		return askErr
	}
	maxPct, parseErr := strconv.ParseFloat(strings.TrimSpace(maxStr), 64)
	if parseErr != nil || maxPct < 0 {
		return fmt.Errorf("invalid max_regression_percent: %s", maxStr)
	}
	p.MaxRegressionPercent = maxPct

	if askErr := survey.AskOne(&survey.Confirm{
		Message: "Fail when a function gets unexpectedly faster (unusual — usually leave off)?",
		Default: p.FailOnImprovement,
		Help:    "Most projects leave this off. Enable only if speedups should also break CI.",
	}, &p.FailOnImprovement); askErr != nil {
		return askErr
	}
	return nil
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func confirmSaveIfChanged(path string, loadedAt time.Time) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.ModTime().After(loadedAt) {
		return nil
	}
	confirm := false
	if askErr := survey.AskOne(&survey.Confirm{
		Message: "prof.json was modified on disk since this session started. Overwrite?",
		Default: false,
	}, &confirm); askErr != nil {
		return askErr
	}
	if !confirm {
		return errors.New("save cancelled: file changed on disk")
	}
	return nil
}

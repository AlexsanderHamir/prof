package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
)

func printCollectionFilterPreview(svc *app.Services, benchNames []string) {
	cfg, err := svc.Config.Load()
	if err != nil {
		fmt.Fprintln(os.Stdout, "Collection filters: none (no prof.json or load failed)")
		return
	}
	fmt.Fprintln(os.Stdout, "Collection filters from prof.json:")
	for _, bench := range benchNames {
		filter := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto(bench))
		fmt.Fprintf(os.Stdout, "  %s — include: %s; ignore: %s\n",
			bench, formatFilterList(filter.IncludePrefixes), formatFilterList(filter.IgnoreFunctions))
	}
}

func formatFilterList(items []string) string {
	if len(items) == 0 {
		return "(all)"
	}
	return strings.Join(items, ", ")
}

func printCollectOutputOptionsHelp() {
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Collect output options:")
	fmt.Fprintln(os.Stdout, "  Lenient profiles — if a .prof file is missing after the bench, skip it instead of failing the run")
	fmt.Fprintln(os.Stdout, "  Skip PNG — ignore call-graph PNG failures; the run still succeeds when text profiles were collected")
	fmt.Fprintln(os.Stdout, "")
	if tooling.GraphvizAvailable() {
		fmt.Fprintln(os.Stdout, "Press Enter for defaults: flat text profiles, strict profile checks, PNG call graphs when possible.")
	} else {
		fmt.Fprintln(os.Stdout, "Press Enter for defaults: flat text profiles, strict profile checks, PNG skipped (Graphviz not installed).")
	}
}

func askAdvancedCollectOptions() (lenient, skipPng bool, err error) {
	printCollectOutputOptionsHelp()

	var advanced bool
	if err = survey.AskOne(&survey.Confirm{
		Message: "Change any of the options above?",
		Default: false,
	}, &advanced); err != nil {
		return false, false, err
	}
	if !advanced {
		return false, !tooling.GraphvizAvailable(), nil
	}
	if err = survey.AskOne(&survey.Confirm{
		Message: "Lenient profiles: skip missing .prof files instead of failing the run?",
		Default: false,
	}, &lenient); err != nil {
		return false, false, err
	}
	skipDefault := !tooling.GraphvizAvailable()
	if err = survey.AskOne(&survey.Confirm{
		Message: "Skip PNG: succeed even when call-graph PNG generation fails?",
		Default: skipDefault,
	}, &skipPng); err != nil {
		return false, false, err
	}
	if !skipPng && !tooling.GraphvizAvailable() {
		fmt.Fprintln(os.Stdout, tooling.SkipPNGNotice)
		skipPng = true
	}
	return lenient, skipPng, nil
}

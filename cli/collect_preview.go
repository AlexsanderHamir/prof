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

func askAdvancedCollectOptions() (groupPkg, lenient, skipPng bool, err error) {
	var advanced bool
	if err = survey.AskOne(&survey.Confirm{
		Message: "Advanced options (group by package, lenient profiles, skip PNG)?",
		Default: false,
	}, &advanced); err != nil {
		return false, false, false, err
	}
	if !advanced {
		return false, false, !tooling.GraphvizAvailable(), nil
	}
	if err = survey.AskOne(&survey.Confirm{
		Message: "Group profile output by Go package (writes additional *_grouped.txt style reports under the tag)?",
		Default: false,
	}, &groupPkg); err != nil {
		return false, false, false, err
	}
	if err = survey.AskOne(&survey.Confirm{
		Message: "Lenient profiles: if a profile binary is missing after the bench run, skip it instead of failing?",
		Default: false,
	}, &lenient); err != nil {
		return false, false, false, err
	}
	skipDefault := !tooling.GraphvizAvailable()
	if err = survey.AskOne(&survey.Confirm{
		Message: "Skip PNG generation failures (e.g. when Graphviz is not installed)? The run still succeeds if text profiles were produced.",
		Default: skipDefault,
	}, &skipPng); err != nil {
		return false, false, false, err
	}
	if !skipPng && !tooling.GraphvizAvailable() {
		fmt.Fprintln(os.Stdout, tooling.SkipPNGNotice)
		skipPng = true
	}
	return groupPkg, lenient, skipPng, nil
}

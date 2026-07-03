package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/engine/tooling"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/termui"
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
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, termui.PromptSectionStyle.Render("Output options"))
	fmt.Fprintln(os.Stdout, termui.PromptHintStyle.Render("  Optional — press Enter at the prompt below to keep defaults."))
	fmt.Fprintln(os.Stdout)

	fmt.Fprintln(os.Stdout, "  Lenient profiles")
	fmt.Fprintln(os.Stdout, termui.PromptHintStyle.Render("    Skip missing .prof files instead of failing the run"))
	fmt.Fprintln(os.Stdout, "  Skip PNG")
	fmt.Fprintln(os.Stdout, termui.PromptHintStyle.Render("    Ignore call-graph PNG failures; text profiles still count"))
	fmt.Fprintln(os.Stdout)

	defaults := "flat text profiles · strict profile checks · PNG call graphs when Graphviz is available"
	if !tooling.GraphvizAvailable() {
		defaults = "flat text profiles · strict profile checks · PNG skipped"
	}
	fmt.Fprintln(os.Stdout, termui.PromptHintStyle.Render("  Defaults: "+defaults))
	fmt.Fprintln(os.Stdout)

	if !tooling.GraphvizAvailable() {
		fmt.Fprintln(os.Stdout, termui.PromptCalloutStyle.Render("Graphviz not installed — PNG call graphs will be skipped."))
		fmt.Fprintln(os.Stdout)
	}
}

func askAdvancedCollectOptions() (lenient, skipPng bool, err error) {
	printCollectOutputOptionsHelp()

	var advanced bool
	if err = survey.AskOne(&survey.Confirm{
		Message: "Customize output options?",
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

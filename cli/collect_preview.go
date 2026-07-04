package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/termui"
)

const missingConfigFilterWarning = "No prof.json found; proceeding without function filters (run prof config init to add one)."

// printCollectionFilterPreview reports filter settings after benchmark selection.
// It returns true when a warning was printed (TTY only); gaps are added around warnings.
func printCollectionFilterPreview(w io.Writer, interactive bool, svc *app.Services, benchNames []string) bool {
	cfg, err := svc.Config.Load()
	if err != nil {
		if interactive {
			termui.PrintWarning(w, termui.ConfigureWarningPrefix, missingConfigFilterWarning)
			termui.StepGap(w)
			return true
		}
		fmt.Fprintln(w, "Collection filters: none (no prof.json or load failed)")
		return false
	}

	if interactive {
		return false
	}

	fmt.Fprintln(w, "Collection filters from prof.json:")
	for _, bench := range benchNames {
		filter := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto(bench))
		fmt.Fprintf(w, "  %s — include: %s; ignore: %s\n",
			bench, formatFilterList(filter.IncludePrefixes), formatFilterList(filter.IgnoreFunctions))
	}
	return false
}

func formatFilterList(items []string) string {
	if len(items) == 0 {
		return "(all)"
	}
	return strings.Join(items, ", ")
}

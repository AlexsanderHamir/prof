package cli

import (
	"fmt"
	"os"
	"strings"

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

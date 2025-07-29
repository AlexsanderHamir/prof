package cli

import (
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"os"
	"sort"

	"github.com/AlexsanderHamir/prof/engine/benchmark"
	"github.com/AlexsanderHamir/prof/engine/tracker"
	"github.com/AlexsanderHamir/prof/internal/args"
	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/shared"
	"github.com/microcosm-cc/bluemonday"
)

func printConfiguration(benchArgs *args.BenchArgs, functionFilterPerBench map[string]config.FunctionFilter) {
	slog.Info(
		"Parsed arguments",
		"Benchmarks", benchArgs.Benchmarks,
		"Profiles", benchArgs.Profiles,
		"Tag", benchArgs.Tag,
		"Count", benchArgs.Count,
	)

	hasBenchFunctionFilters := len(functionFilterPerBench) > 0
	if hasBenchFunctionFilters {
		slog.Info("Benchmark Function Filter Configurations:")
		for benchmark, cfg := range functionFilterPerBench {
			slog.Info("Benchmark Config", "Benchmark", benchmark, "Prefixes", cfg.IncludePrefixes, "Ignore", cfg.IgnoreFunctions)
		}
	} else {
		slog.Info("No benchmark configuration found in config file - analyzing all functions")
	}
}

func runBenchAndGetProfiles(benchArgs *args.BenchArgs, benchmarkConfigs map[string]config.FunctionFilter) error {
	slog.Info("Starting benchmark pipeline...")

	var functionFilter config.FunctionFilter
	globalFilter, hasGlobalFilter := benchmarkConfigs[shared.GlobalSign]
	if hasGlobalFilter {
		functionFilter = globalFilter
	}

	for _, benchmarkName := range benchArgs.Benchmarks {
		slog.Info("Running benchmark", "Benchmark", benchmarkName)
		if err := benchmark.RunBenchmark(benchmarkName, benchArgs.Profiles, benchArgs.Count, benchArgs.Tag); err != nil {
			return fmt.Errorf("failed to run %s: %w", benchmarkName, err)
		}

		slog.Info("Processing profiles", "Benchmark", benchmarkName)
		if err := benchmark.ProcessProfiles(benchmarkName, benchArgs.Profiles, benchArgs.Tag); err != nil {
			return fmt.Errorf("failed to process profiles for %s: %w", benchmarkName, err)
		}

		slog.Info("Analyzing profile functions", "Benchmark", benchmarkName)

		if !hasGlobalFilter {
			functionFilter = benchmarkConfigs[benchmarkName]
		}

		args := &args.CollectionArgs{
			Tag:             benchArgs.Tag,
			Profiles:        benchArgs.Profiles,
			BenchmarkName:   benchmarkName,
			BenchmarkConfig: functionFilter,
		}

		if err := benchmark.CollectProfileFunctions(args); err != nil {
			return fmt.Errorf("failed to analyze profile functions for %s: %w", benchmarkName, err)
		}

		slog.Info("Completed pipeline for benchmark", "Benchmark", benchmarkName)
	}

	slog.Info(shared.InfoCollectionSuccess)
	return nil
}

func printSummary(report *tracker.ProfileChangeReport) {
	fmt.Println("\n=== Performance Tracking Summary ===")

	var regressionList, improvementList []*tracker.FunctionChangeResult
	var stable int

	// Separate changes by type
	for _, change := range report.FunctionChanges {
		switch change.ChangeType {
		case shared.REGRESSION:
			regressionList = append(regressionList, change)
		case shared.IMPROVEMENT:
			improvementList = append(improvementList, change)
		default:
			stable++
		}
	}

	// Sort regressions by percentage (biggest regression first)
	sort.Slice(regressionList, func(i, j int) bool {
		return regressionList[i].FlatChangePercent > regressionList[j].FlatChangePercent
	})

	// Sort improvements by absolute percentage (biggest improvement first)
	sort.Slice(improvementList, func(i, j int) bool {
		return math.Abs(improvementList[i].FlatChangePercent) > math.Abs(improvementList[j].FlatChangePercent)
	})

	fmt.Printf("Total Functions Analyzed: %d\n", len(report.FunctionChanges))
	fmt.Printf("Regressions: %d\n", len(regressionList))
	fmt.Printf("Improvements: %d\n", len(improvementList))
	fmt.Printf("Stable: %d\n", stable)

	if len(regressionList) > 0 {
		fmt.Println("\n⚠️  Top Regressions (worst first):")
		for _, change := range regressionList {
			fmt.Printf("  • %s\n", change.Summary())
		}
	}

	if len(improvementList) > 0 {
		fmt.Println("\n✅ Top Improvements (best first):")
		for _, change := range improvementList {
			fmt.Printf("  • %s\n", change.Summary())
		}
	}
}

const (
	regressionPriority  = 1
	improvementPriority = 2
	stablePriority      = 3
)

func printDetailedReport(report *tracker.ProfileChangeReport) {
	changes := report.FunctionChanges

	// Count each type
	var regressions, improvements, stable int
	for _, change := range changes {
		switch change.ChangeType {
		case shared.REGRESSION:
			regressions++
		case shared.IMPROVEMENT:
			improvements++
		default:
			stable++
		}
	}

	// Print header with statistics and sorting info
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Detailed Performance Report                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n📊 Summary: %d total functions | 🔴 %d regressions | 🟢 %d improvements | ⚪ %d stable\n",
		len(changes), regressions, improvements, stable)
	fmt.Println("\n📋 Report Order: Regressions first (worst → best), then Improvements (best → worst), then Stable")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════")

	// Sort by change type first (REGRESSION, IMPROVEMENT, STABLE),
	// then by absolute percentage change (biggest changes first)
	sort.Slice(changes, func(i, j int) bool {
		// Primary sort: by change type priority
		typePriority := map[string]int{
			shared.REGRESSION:  regressionPriority,
			shared.IMPROVEMENT: improvementPriority,
			shared.STABLE:      stablePriority,
		}

		if typePriority[changes[i].ChangeType] != typePriority[changes[j].ChangeType] {
			return typePriority[changes[i].ChangeType] < typePriority[changes[j].ChangeType]
		}

		return math.Abs(changes[i].FlatChangePercent) > math.Abs(changes[j].FlatChangePercent)
	})

	for i, change := range changes {
		if i > 0 {
			fmt.Println()
			fmt.Println()
			fmt.Println()
		}
		fmt.Print(change.Report())
	}
}

type htmlData struct {
	TotalFunctions int
	Regressions    []*tracker.FunctionChangeResult
	Improvements   []*tracker.FunctionChangeResult
	Stable         int
}

func generateHTMLSummary(report *tracker.ProfileChangeReport, outputPath string) error {
	var regressionList, improvementList []*tracker.FunctionChangeResult
	var stable int

	for _, change := range report.FunctionChanges {
		switch change.ChangeType {
		case shared.REGRESSION:
			regressionList = append(regressionList, change)
		case shared.IMPROVEMENT:
			improvementList = append(improvementList, change)
		default:
			stable++
		}
	}

	sort.Slice(regressionList, func(i, j int) bool {
		return regressionList[i].FlatChangePercent > regressionList[j].FlatChangePercent
	})
	sort.Slice(improvementList, func(i, j int) bool {
		return math.Abs(improvementList[i].FlatChangePercent) > math.Abs(improvementList[j].FlatChangePercent)
	})

	data := htmlData{
		TotalFunctions: len(report.FunctionChanges),
		Regressions:    regressionList,
		Improvements:   improvementList,
		Stable:         stable,
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Performance Tracking Summary</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			margin: 2rem;
			font-size: 1.25rem; /* Increased again */
		}
		h2 {
			color: #333;
			font-size: 1.8rem; /* Bigger main heading */
		}
		h3 {
			font-size: 1.5rem; /* Bigger subheading */
		}
		.regressions { color: red; }
		.improvements { color: green; }
		ul { padding-left: 1.5rem; }
	</style>
</head>
<body>
	<h2>Performance Tracking Summary</h2>
	<p><strong>Total Functions Analyzed:</strong> {{.TotalFunctions}}</p>
	<p><strong>Regressions:</strong> {{len .Regressions}}</p>
	<p><strong>Improvements:</strong> {{len .Improvements}}</p>
	<p><strong>Stable:</strong> {{.Stable}}</p>

	{{if .Regressions}}
		<h3 class="regressions">⚠️ Top Regressions (worst first):</h3>
		<ul>
			{{range .Regressions}}<li>{{summary .}}</li>{{end}}
		</ul>
	{{end}}

	{{if .Improvements}}
		<h3 class="improvements">✅ Top Improvements (best first):</h3>
		<ul>
			{{range .Improvements}}<li>{{summary .}}</li>{{end}}
		</ul>
	{{end}}
</body>
</html>
`

	funcMap := template.FuncMap{
		"summary": func(fc *tracker.FunctionChangeResult) string {
			return fc.Summary()
		},
	}

	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

type detailedHTMLData struct {
	Total        int
	Regressions  int
	Improvements int
	Stable       int
	Changes      []*tracker.FunctionChangeResult
}

func generateDetailedHTMLReport(report *tracker.ProfileChangeReport, outputPath string) error {
	changes := report.FunctionChanges

	// Count types
	var regressions, improvements, stable int
	for _, change := range changes {
		switch change.ChangeType {
		case shared.REGRESSION:
			regressions++
		case shared.IMPROVEMENT:
			improvements++
		default:
			stable++
		}
	}

	// Sort: regressions → improvements → stable, each by magnitude
	typePriority := map[string]int{
		shared.REGRESSION:  regressionPriority,
		shared.IMPROVEMENT: improvementPriority,
		shared.STABLE:      stablePriority,
	}

	sort.Slice(changes, func(i, j int) bool {
		if typePriority[changes[i].ChangeType] != typePriority[changes[j].ChangeType] {
			return typePriority[changes[i].ChangeType] < typePriority[changes[j].ChangeType]
		}
		return math.Abs(changes[i].FlatChangePercent) > math.Abs(changes[j].FlatChangePercent)
	})

	data := detailedHTMLData{
		Total:        len(changes),
		Regressions:  regressions,
		Improvements: improvements,
		Stable:       stable,
		Changes:      changes,
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Detailed Performance Report</title>
	<style>
		body { font-family: monospace, sans-serif; padding: 2rem; background: #f8f8f8; }
		h1, h2 { color: #333; }
		.stats { margin-bottom: 1rem; font-family: sans-serif; }
		.report-block { margin-bottom: 3rem; white-space: pre-wrap; 
		background: #fff; padding: 1rem; border-left: 4px solid #ccc; 
		font-size: 1.25rem; }
	</style>
</head>
<body>
	<h1>📈 Detailed Performance Report</h1>

	<div class="stats">
		<p><strong>Total functions:</strong> {{.Total}}</p>
		<p><strong>🔴 Regressions:</strong> {{.Regressions}} | <strong>🟢 Improvements:</strong> {{.Improvements}} | <strong>⚪ Stable:</strong> {{.Stable}}</p>
		<p><em>Report Order: Regressions (worst → best), then Improvements (best → worst), then Stable</em></p>
	</div>

	{{range .Changes}}
		<div class="report-block">
			<pre>{{report .}}</pre>
		</div>
	{{end}}
</body>
</html>
`

	sanitizer := bluemonday.StrictPolicy()
	funcMap := template.FuncMap{
		"report": func(fc *tracker.FunctionChangeResult) template.HTML {
			safe := sanitizer.Sanitize(fc.Report())
			return template.HTML(safe) //nolint:gosec // input is being sanatized
		},
	}

	t, err := template.New("detailed").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

func chooseOutputFormat(report *tracker.ProfileChangeReport) {
	switch outputFormat {
	case "summary":
		printSummary(report)
	case "detailed":
		printDetailedReport(report)
	case "summary-html":
		if err := generateHTMLSummary(report, "summary.html"); err != nil {
			slog.Info("summary-html failed", "err", err)
		}
	case "detailed-html":
		if err := generateDetailedHTMLReport(report, "detailed.html"); err != nil {
			slog.Info("detailed-html failed", "err", err)
		}
	}
}

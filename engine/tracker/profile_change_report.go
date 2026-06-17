package tracker

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"os"
	"sort"

	"github.com/AlexsanderHamir/prof/internal/config"

	"github.com/microcosm-cc/bluemonday"
)

func (r *ProfileChangeReport) printSummary() {
	fmt.Println("\n=== Performance Tracking Summary ===")

	var regressionList, improvementList []*FunctionChangeResult
	var stable int

	// Separate changes by type
	for _, change := range r.FunctionChanges {
		switch change.ChangeType {
		case ChangeRegression:
			regressionList = append(regressionList, change)
		case ChangeImprovement:
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

	fmt.Printf("Total Functions Analyzed: %d\n", len(r.FunctionChanges))
	fmt.Printf("Regressions: %d\n", len(regressionList))
	fmt.Printf("Improvements: %d\n", len(improvementList))
	fmt.Printf("Stable: %d\n", stable)

	if len(regressionList) > 0 {
		fmt.Println("\n⚠️  Top Regressions (worst first):")
		for _, change := range regressionList {
			fmt.Printf("  • %s\n", change.summary())
		}
	}

	if len(improvementList) > 0 {
		fmt.Println("\n✅ Top Improvements (best first):")
		for _, change := range improvementList {
			fmt.Printf("  • %s\n", change.summary())
		}
	}
}

const (
	regressionPriority  = 1
	improvementPriority = 2
	stablePriority      = 3
)

func (r *ProfileChangeReport) printDetailedReport() {
	changes := r.FunctionChanges

	// Count each type
	var regressions, improvements, stable int
	for _, change := range changes {
		switch change.ChangeType {
		case ChangeRegression:
			regressions++
		case ChangeImprovement:
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
			ChangeRegression:  regressionPriority,
			ChangeImprovement: improvementPriority,
			ChangeStable:      stablePriority,
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
	Regressions    []*FunctionChangeResult
	Improvements   []*FunctionChangeResult
	Stable         int
}

func (r *ProfileChangeReport) generateHTMLSummary(outputPath string) error {
	var regressionList, improvementList []*FunctionChangeResult
	var stable int

	for _, change := range r.FunctionChanges {
		switch change.ChangeType {
		case ChangeRegression:
			regressionList = append(regressionList, change)
		case ChangeImprovement:
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
		TotalFunctions: len(r.FunctionChanges),
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
		"summary": func(fc *FunctionChangeResult) string {
			return fc.summary()
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
	Changes      []*FunctionChangeResult
}

func (r *ProfileChangeReport) generateDetailedHTMLReport(outputPath string) error {
	changes := r.FunctionChanges

	// Count types
	var regressions, improvements, stable int
	for _, change := range changes {
		switch change.ChangeType {
		case ChangeRegression:
			regressions++
		case ChangeImprovement:
			improvements++
		default:
			stable++
		}
	}

	// Sort: regressions → improvements → stable, each by magnitude
	typePriority := map[string]int{
		ChangeRegression:  regressionPriority,
		ChangeImprovement: improvementPriority,
		ChangeStable:      stablePriority,
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
		"report": func(fc *FunctionChangeResult) template.HTML {
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

// JSON data structures
type jsonSummaryData struct {
	TotalFunctions int                     `json:"total_functions"`
	Statistics     jsonStatistics          `json:"statistics"`
	Regressions    []*FunctionChangeResult `json:"regressions"`
	Improvements   []*FunctionChangeResult `json:"improvements"`
}

type jsonStatistics struct {
	Regressions  int `json:"regressions"`
	Improvements int `json:"improvements"`
	Stable       int `json:"stable"`
}

type jsonDetailedData struct {
	TotalFunctions int                     `json:"total_functions"`
	Statistics     jsonStatistics          `json:"statistics"`
	Changes        []*FunctionChangeResult `json:"changes"`
	SortOrder      string                  `json:"sort_order"`
}

func (r *ProfileChangeReport) generateJSONSummary(outputPath string) error {
	var regressionList, improvementList []*FunctionChangeResult
	var stable int

	for _, change := range r.FunctionChanges {
		switch change.ChangeType {
		case ChangeRegression:
			regressionList = append(regressionList, change)
		case ChangeImprovement:
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

	data := jsonSummaryData{
		TotalFunctions: len(r.FunctionChanges),
		Statistics: jsonStatistics{
			Regressions:  len(regressionList),
			Improvements: len(improvementList),
			Stable:       stable,
		},
		Regressions:  regressionList,
		Improvements: improvementList,
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (r *ProfileChangeReport) generateDetailedJSONReport(outputPath string) error {
	changes := r.FunctionChanges

	// Count types
	var regressions, improvements, stable int
	for _, change := range changes {
		switch change.ChangeType {
		case ChangeRegression:
			regressions++
		case ChangeImprovement:
			improvements++
		default:
			stable++
		}
	}

	// Sort: regressions → improvements → stable, each by magnitude
	typePriority := map[string]int{
		ChangeRegression:  regressionPriority,
		ChangeImprovement: improvementPriority,
		ChangeStable:      stablePriority,
	}

	sort.Slice(changes, func(i, j int) bool {
		if typePriority[changes[i].ChangeType] != typePriority[changes[j].ChangeType] {
			return typePriority[changes[i].ChangeType] < typePriority[changes[j].ChangeType]
		}
		return math.Abs(changes[i].FlatChangePercent) > math.Abs(changes[j].FlatChangePercent)
	})

	data := jsonDetailedData{
		TotalFunctions: len(changes),
		Statistics: jsonStatistics{
			Regressions:  regressions,
			Improvements: improvements,
			Stable:       stable,
		},
		Changes:   changes,
		SortOrder: "Regressions (worst → best), then Improvements (best → worst), then Stable",
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ChooseOutputFormat dispatches to the appropriate report formatter for the given output name.
func (r *ProfileChangeReport) ChooseOutputFormat(outputFormat string) error {
	switch outputFormat {
	case "summary":
		r.printSummary()
		return nil
	case "detailed":
		r.printDetailedReport()
		return nil
	case "summary-html":
		return r.generateHTMLSummary("summary.html")
	case "detailed-html":
		return r.generateDetailedHTMLReport("detailed.html")
	case "summary-json":
		return r.generateJSONSummary("summary.json")
	case "detailed-json":
		return r.generateDetailedJSONReport("detailed.json")
	default:
		return fmt.Errorf("unsupported output format %q", outputFormat)
	}
}

// WorstRegression returns the single worst regression by flat change percent.
// Returns nil if there are no regressions.
func (r *ProfileChangeReport) WorstRegression() *FunctionChangeResult {
	var worst *FunctionChangeResult
	for _, change := range r.FunctionChanges {
		if change.ChangeType != ChangeRegression {
			continue
		}
		if worst == nil || change.FlatChangePercent > worst.FlatChangePercent {
			worst = change
		}
	}
	return worst
}

// BestImprovement returns the single best improvement by absolute flat change percent.
// Returns nil if there are no improvements.
func (r *ProfileChangeReport) BestImprovement() *FunctionChangeResult {
	var best *FunctionChangeResult
	for _, change := range r.FunctionChanges {
		if change.ChangeType != ChangeImprovement {
			continue
		}
		if best == nil || math.Abs(change.FlatChangePercent) > math.Abs(best.FlatChangePercent) {
			best = change
		}
	}
	return best
}

// ApplyTrackPolicy filters ignored functions from the report using track policy.
func (r *ProfileChangeReport) ApplyTrackPolicy(policy config.TrackPolicy) {
	originalCount := len(r.FunctionChanges)
	var filteredChanges []*FunctionChangeResult
	for _, change := range r.FunctionChanges {
		if !config.ShouldIgnoreFunction(policy, change.FunctionName) {
			filteredChanges = append(filteredChanges, change)
		} else {
			slog.Debug("Function ignored by track config", "function", change.FunctionName)
		}
	}

	r.FunctionChanges = filteredChanges

	if originalCount != len(filteredChanges) {
		slog.Info("Applied track configuration filtering",
			"original_functions", originalCount,
			"filtered_functions", len(filteredChanges))
	}
}

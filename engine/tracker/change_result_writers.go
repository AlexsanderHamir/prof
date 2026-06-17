package tracker

import (
	"fmt"
	"math"
	"strings"
)

func (cr *FunctionChangeResult) writeHeader(report *strings.Builder) {
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString("               PERFORMANCE CHANGE REPORT\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
}

func (cr *FunctionChangeResult) writeFunctionInfo(report *strings.Builder) {
	fmt.Fprintf(report, "Function: %s\n", cr.FunctionName)
	fmt.Fprintf(report, "Analysis Time: %s\n", cr.Timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(report, "Change Type: %s\n\n", cr.ChangeType)
}

func (cr *FunctionChangeResult) writeStatusAssessment(report *strings.Builder) {
	statusIcon := map[string]string{
		ChangeImprovement: "✅",
		ChangeRegression:  "⚠️",
	}[cr.ChangeType]
	if statusIcon == "" {
		statusIcon = "🔄"
	}
	assessment := map[string]string{
		ChangeImprovement: "Performance improvement detected",
		ChangeRegression:  "Performance regression detected",
	}[cr.ChangeType]
	if assessment == "" {
		assessment = "No significant change detected"
	}
	fmt.Fprintf(report, "%s %s\n\n", statusIcon, assessment)
}

func (cr *FunctionChangeResult) writeFlatAnalysis(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    FLAT TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	sign := signPrefix(cr.FlatChangePercent)
	fmt.Fprintf(report, "Before:       %.6fs\n", cr.FlatAbsolute.Before)
	fmt.Fprintf(report, "After:        %.6fs\n", cr.FlatAbsolute.After)
	fmt.Fprintf(report, "Delta:        %s%.6fs\n", sign, cr.FlatAbsolute.Delta)
	fmt.Fprintf(report, "Change:       %s%.2f%%\n", sign, cr.FlatChangePercent)
	switch {
	case cr.FlatChangePercent > 0:
		fmt.Fprintf(report, "Impact:       Function is %.2f%% SLOWER\n\n", cr.FlatChangePercent)
	case cr.FlatChangePercent < 0:
		fmt.Fprintf(report, "Impact:       Function is %.2f%% FASTER\n\n", math.Abs(cr.FlatChangePercent))
	default:
		report.WriteString("Impact:       No change in execution time\n\n")
	}
}

func (cr *FunctionChangeResult) writeCumulativeAnalysis(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                 CUMULATIVE TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	sign := signPrefix(cr.CumChangePercent)
	fmt.Fprintf(report, "Before:       %.3fs\n", cr.CumAbsolute.Before)
	fmt.Fprintf(report, "After:        %.3fs\n", cr.CumAbsolute.After)
	fmt.Fprintf(report, "Delta:        %s%.3fs\n", sign, cr.CumAbsolute.Delta)
	fmt.Fprintf(report, "Change:       %s%.2f%%\n\n", sign, cr.CumChangePercent)
}

func (cr *FunctionChangeResult) writeImpactAssessment(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    IMPACT ASSESSMENT\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(report, "Severity:     %s\n", cr.calculateSeverity())
	report.WriteString("Recommendation: ")
	report.WriteString(cr.recommendation())
	report.WriteString("\n")
}

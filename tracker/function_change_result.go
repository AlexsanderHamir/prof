package tracker

import (
	"fmt"
	"math"
	"strings"
)

// Helper method to get a summary
func (cr *FunctionChangeResult) Summary() string {
	sign := ""
	if cr.FlatChangePercent > 0 {
		sign = "+"
	}

	return fmt.Sprintf("%s: %s%.1f%% (%.3fs → %.3fs)",
		cr.ChangeType,
		sign,
		cr.FlatChangePercent,
		cr.FlatAbsolute.Before,
		cr.FlatAbsolute.After)
}

// Full detailed report
func (cr *FunctionChangeResult) Report() string {
	var report strings.Builder

	// Header
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString("               PERFORMANCE CHANGE REPORT\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Function Information
	report.WriteString(fmt.Sprintf("Function: %s\n", cr.FunctionName))
	report.WriteString(fmt.Sprintf("Analysis Time: %s\n", cr.Timestamp.Format("2006-01-02 15:04:05 MST")))
	report.WriteString(fmt.Sprintf("Change Type: %s\n", cr.ChangeType))
	report.WriteString("\n")

	// Status Icon and Overall Assessment
	statusIcon := "🔄"
	assessment := "No significant change detected"

	switch cr.ChangeType {
	case "IMPROVEMENT":
		statusIcon = "✅"
		assessment = "Performance improvement detected"
	case "REGRESSION":
		statusIcon = "⚠️"
		assessment = "Performance regression detected"
	}

	report.WriteString(fmt.Sprintf("%s %s\n", statusIcon, assessment))
	report.WriteString("\n")

	// Flat Time Analysis (Primary Metric)
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    FLAT TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	flatSign := ""
	if cr.FlatChangePercent > 0 {
		flatSign = "+"
	}

	report.WriteString(fmt.Sprintf("Before:       %.6fs\n", cr.FlatAbsolute.Before))
	report.WriteString(fmt.Sprintf("After:        %.6fs\n", cr.FlatAbsolute.After))
	report.WriteString(fmt.Sprintf("Delta:        %s%.6fs\n", flatSign, cr.FlatAbsolute.Delta))
	report.WriteString(fmt.Sprintf("Change:       %s%.2f%%\n", flatSign, cr.FlatChangePercent))

	// Interpretation
	if cr.FlatChangePercent > 0 {
		report.WriteString(fmt.Sprintf("Impact:       Function is %.2f%% SLOWER\n", cr.FlatChangePercent))
	} else if cr.FlatChangePercent < 0 {
		report.WriteString(fmt.Sprintf("Impact:       Function is %.2f%% FASTER\n", math.Abs(cr.FlatChangePercent)))
	} else {
		report.WriteString("Impact:       No change in execution time\n")
	}
	report.WriteString("\n")

	// Cumulative Time Analysis (Secondary Metric)
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                 CUMULATIVE TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	cumSign := ""
	if cr.CumChangePercent > 0 {
		cumSign = "+"
	}

	report.WriteString(fmt.Sprintf("Before:       %.3fs\n", cr.CumAbsolute.Before))
	report.WriteString(fmt.Sprintf("After:        %.3fs\n", cr.CumAbsolute.After))
	report.WriteString(fmt.Sprintf("Delta:        %s%.3fs\n", cumSign, cr.CumAbsolute.Delta))
	report.WriteString(fmt.Sprintf("Change:       %s%.2f%%\n", cumSign, cr.CumChangePercent))

	// Overall Impact Assessment
	report.WriteString("\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    IMPACT ASSESSMENT\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	// Severity classification
	severity := cr.calculateSeverity()
	report.WriteString(fmt.Sprintf("Severity:     %s\n", severity))

	// Recommendations
	report.WriteString("Recommendation: ")
	switch cr.ChangeType {
	case "IMPROVEMENT":
		if math.Abs(cr.FlatChangePercent) > 25 {
			report.WriteString("Significant performance gain! Consider documenting the optimization.\n")
		} else if math.Abs(cr.FlatChangePercent) > 10 {
			report.WriteString("Notable improvement detected. Monitor to ensure consistency.\n")
		} else {
			report.WriteString("Minor improvement detected. Continue monitoring.\n")
		}
	case "REGRESSION":
		if cr.FlatChangePercent > 50 {
			report.WriteString("Critical regression! Immediate investigation required.\n")
		} else if cr.FlatChangePercent > 25 {
			report.WriteString("Significant regression detected. Consider rollback or optimization.\n")
		} else if cr.FlatChangePercent > 10 {
			report.WriteString("Moderate regression. Review recent changes and optimize if needed.\n")
		} else {
			report.WriteString("Minor regression detected. Monitor for trends.\n")
		}
	default:
		report.WriteString("No action required. Continue monitoring.\n")
	}

	report.WriteString("\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")

	return report.String()
}

// Helper method to calculate severity
func (cr *FunctionChangeResult) calculateSeverity() string {
	absChange := math.Abs(cr.FlatChangePercent)

	if absChange == 0 {
		return "NONE"
	} else if absChange < 5 {
		return "LOW"
	} else if absChange < 15 {
		return "MODERATE"
	} else if absChange < 30 {
		return "HIGH"
	} else {
		return "CRITICAL"
	}
}

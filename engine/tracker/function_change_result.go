package tracker

import (
	"fmt"
	"math"
	"strings"
)

// Helper method to get a summary
func (cr *FunctionChangeResult) summary() string {
	sign := ""
	if cr.FlatChangePercent > 0 {
		sign = "+"
	}

	return fmt.Sprintf("%s: %s%.1f%% (%.3fs → %.3fs)",
		cr.FunctionName,
		sign,
		cr.FlatChangePercent,
		cr.FlatAbsolute.Before,
		cr.FlatAbsolute.After)
}

// Full detailed report
func (cr *FunctionChangeResult) Report() string {
	var report strings.Builder

	cr.writeHeader(&report)
	cr.writeFunctionInfo(&report)
	cr.writeStatusAssessment(&report)
	cr.writeFlatAnalysis(&report)
	cr.writeCumulativeAnalysis(&report)
	cr.writeImpactAssessment(&report)

	report.WriteString("\n═══════════════════════════════════════════════════════════════\n")
	return report.String()
}

const (
	SeverityNoneThreshold     = 0.0
	SeverityLowThreshold      = 5.0
	SeverityModerateThreshold = 15.0
	SeverityHighThreshold     = 30.0
)

func (cr *FunctionChangeResult) calculateSeverity() string {
	absChange := math.Abs(cr.FlatChangePercent)

	switch {
	case absChange == SeverityNoneThreshold:
		return "NONE"
	case absChange < SeverityLowThreshold:
		return "LOW"
	case absChange < SeverityModerateThreshold:
		return "MODERATE"
	case absChange < SeverityHighThreshold:
		return "HIGH"
	default:
		return "CRITICAL"
	}
}

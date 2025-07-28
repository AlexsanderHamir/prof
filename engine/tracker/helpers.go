package tracker

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AlexsanderHamir/prof/internal/shared"
	"github.com/AlexsanderHamir/prof/parser"
)

func createHashFromLineObjects(lineobjects []*parser.LineObj) map[string]*parser.LineObj {
	matchingMap := make(map[string]*parser.LineObj)
	for _, lineObj := range lineobjects {
		matchingMap[lineObj.FnName] = lineObj
	}

	return matchingMap
}

func DetectChange(baseline, current *parser.LineObj) (*FunctionChangeResult, error) {
	if current == nil {
		return nil, errors.New("current obj is nil")
	}
	if baseline == nil {
		return nil, errors.New("baseLine obj is nil")
	}

	const percentMultiplier = 100

	var flatChange float64
	if baseline.Flat != 0 {
		flatChange = ((current.Flat - baseline.Flat) / baseline.Flat) * percentMultiplier
	}

	var cumChange float64
	if baseline.Cum != 0 {
		cumChange = ((current.Cum - baseline.Cum) / baseline.Cum) * percentMultiplier
	}

	changeType := shared.STABLE
	if flatChange > 0 {
		changeType = shared.REGRESSION
	} else if flatChange < 0 {
		changeType = shared.IMPROVEMENT
	}

	return &FunctionChangeResult{
		FunctionName:      current.FnName,
		ChangeType:        changeType,
		FlatChangePercent: flatChange,
		CumChangePercent:  cumChange,
		FlatAbsolute: struct {
			Before float64
			After  float64
			Delta  float64
		}{
			Before: baseline.Flat,
			After:  current.Flat,
			Delta:  current.Flat - baseline.Flat,
		},
		CumAbsolute: struct {
			Before float64
			After  float64
			Delta  float64
		}{
			Before: baseline.Cum,
			After:  current.Cum,
			Delta:  current.Cum - baseline.Cum,
		},
		Timestamp: time.Now(),
	}, nil
}

func (cr *FunctionChangeResult) writeHeader(report *strings.Builder) {
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString("               PERFORMANCE CHANGE REPORT\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
}

func (cr *FunctionChangeResult) writeFunctionInfo(report *strings.Builder) {
	report.WriteString(fmt.Sprintf("Function: %s\n", cr.FunctionName))
	report.WriteString(fmt.Sprintf("Analysis Time: %s\n", cr.Timestamp.Format("2006-01-02 15:04:05 MST")))
	report.WriteString(fmt.Sprintf("Change Type: %s\n\n", cr.ChangeType))
}

func (cr *FunctionChangeResult) writeStatusAssessment(report *strings.Builder) {
	statusIcon := map[string]string{
		shared.IMPROVEMENT: "✅",
		shared.REGRESSION:  "⚠️",
	}[cr.ChangeType]

	if statusIcon == "" {
		statusIcon = "🔄"
	}

	assessment := map[string]string{
		shared.IMPROVEMENT: "Performance improvement detected",
		shared.REGRESSION:  "Performance regression detected",
	}[cr.ChangeType]

	if assessment == "" {
		assessment = "No significant change detected"
	}

	report.WriteString(fmt.Sprintf("%s %s\n\n", statusIcon, assessment))
}

func (cr *FunctionChangeResult) writeFlatAnalysis(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    FLAT TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	sign := signPrefix(cr.FlatChangePercent)

	report.WriteString(fmt.Sprintf("Before:       %.6fs\n", cr.FlatAbsolute.Before))
	report.WriteString(fmt.Sprintf("After:        %.6fs\n", cr.FlatAbsolute.After))
	report.WriteString(fmt.Sprintf("Delta:        %s%.6fs\n", sign, cr.FlatAbsolute.Delta))
	report.WriteString(fmt.Sprintf("Change:       %s%.2f%%\n", sign, cr.FlatChangePercent))

	switch {
	case cr.FlatChangePercent > 0:
		report.WriteString(fmt.Sprintf("Impact:       Function is %.2f%% SLOWER\n\n", cr.FlatChangePercent))
	case cr.FlatChangePercent < 0:
		report.WriteString(fmt.Sprintf("Impact:       Function is %.2f%% FASTER\n\n", math.Abs(cr.FlatChangePercent)))
	default:
		report.WriteString("Impact:       No change in execution time\n\n")
	}
}

func (cr *FunctionChangeResult) writeCumulativeAnalysis(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                 CUMULATIVE TIME ANALYSIS\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	sign := signPrefix(cr.CumChangePercent)

	report.WriteString(fmt.Sprintf("Before:       %.3fs\n", cr.CumAbsolute.Before))
	report.WriteString(fmt.Sprintf("After:        %.3fs\n", cr.CumAbsolute.After))
	report.WriteString(fmt.Sprintf("Delta:        %s%.3fs\n", sign, cr.CumAbsolute.Delta))
	report.WriteString(fmt.Sprintf("Change:       %s%.2f%%\n\n", sign, cr.CumChangePercent))
}

func (cr *FunctionChangeResult) writeImpactAssessment(report *strings.Builder) {
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	report.WriteString("                    IMPACT ASSESSMENT\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")

	report.WriteString(fmt.Sprintf("Severity:     %s\n", cr.calculateSeverity()))
	report.WriteString("Recommendation: ")
	report.WriteString(cr.recommendation())
	report.WriteString("\n")
}

package tracker

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

func createMapFromLineObjects(lineobjects []*parser.LineObj) map[string]*parser.LineObj {
	matchingMap := make(map[string]*parser.LineObj)
	for _, lineObj := range lineobjects {
		matchingMap[lineObj.FnName] = lineObj
	}

	return matchingMap
}

func detectChangeBetweenTwoObjects(baseline, current *parser.LineObj) (*FunctionChangeResult, error) {
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

	changeType := internal.STABLE
	if flatChange > 0 {
		changeType = internal.REGRESSION
	} else if flatChange < 0 {
		changeType = internal.IMPROVEMENT
	}

	return &FunctionChangeResult{
		FunctionName:      current.FnName,
		ChangeType:        changeType,
		FlatChangePercent: flatChange,
		CumChangePercent:  cumChange,
		FlatAbsolute: AbsoluteChange{
			Before: baseline.Flat,
			After:  current.Flat,
			Delta:  current.Flat - baseline.Flat,
		},
		CumAbsolute: AbsoluteChange{
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
	fmt.Fprintf(report, "Function: %s\n", cr.FunctionName)
	fmt.Fprintf(report, "Analysis Time: %s\n", cr.Timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(report, "Change Type: %s\n\n", cr.ChangeType)
}

func (cr *FunctionChangeResult) writeStatusAssessment(report *strings.Builder) {
	statusIcon := map[string]string{
		internal.IMPROVEMENT: "✅",
		internal.REGRESSION:  "⚠️",
	}[cr.ChangeType]

	if statusIcon == "" {
		statusIcon = "🔄"
	}

	assessment := map[string]string{
		internal.IMPROVEMENT: "Performance improvement detected",
		internal.REGRESSION:  "Performance regression detected",
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

func getBinFilesLocations(selections *Selections) (string, string) {
	fileName := fmt.Sprintf("%s_%s.txt", selections.BenchmarkName, selections.ProfileType)
	binFilePath1BaseLine := filepath.Join(internal.MainDirOutput, selections.Baseline, internal.ProfileBinDir, selections.BenchmarkName, fileName)
	binFilePath2Current := filepath.Join(internal.MainDirOutput, selections.Current, internal.ProfileBinDir, selections.BenchmarkName, fileName)

	return binFilePath1BaseLine, binFilePath2Current
}

func chooseFileLocations(selections *Selections) (string, string) {
	var textFilePathBaseLine, textFilePathCurrent string

	if selections.IsManual {
		textFilePathBaseLine = selections.Baseline
		textFilePathCurrent = selections.Current
	} else {
		textFilePathBaseLine, textFilePathCurrent = getBinFilesLocations(selections)
	}

	return textFilePathBaseLine, textFilePathCurrent
}

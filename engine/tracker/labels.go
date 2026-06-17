package tracker

// Change type labels for profile comparison results.
const (
	ChangeImprovement = "IMPROVEMENT"
	ChangeRegression  = "REGRESSION"
	ChangeStable      = "STABLE"
)

// ValidOutputFormats lists allowed track report formats.
var ValidOutputFormats = []string{
	"summary",
	"detailed",
	"summary-html",
	"detailed-html",
	"summary-json",
	"detailed-json",
}

// ValidOutputFormat reports whether format is supported.
func ValidOutputFormat(format string) bool {
	for _, f := range ValidOutputFormats {
		if f == format {
			return true
		}
	}
	return false
}

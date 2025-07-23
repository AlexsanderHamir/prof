package tracker

import (
	"math"

	"github.com/AlexsanderHamir/prof/shared"
)

const (
	significantThreshold = 25.0
	notableThreshold     = 10.0
	criticalThreshold    = 50.0
	moderateThreshold    = 10.0
)

func signPrefix(val float64) string {
	if val > 0 {
		return "+"
	}
	return ""
}

func (cr *FunctionChangeResult) recommendation() string {
	switch cr.ChangeType {
	case shared.IMPROVEMENT:
		absChange := math.Abs(cr.FlatChangePercent)
		switch {
		case absChange > significantThreshold:
			return "Significant performance gain! Consider documenting the optimization."
		case absChange > notableThreshold:
			return "Notable improvement detected. Monitor to ensure consistency."
		default:
			return "Minor improvement detected. Continue monitoring."
		}
	case shared.REGRESSION:
		switch {
		case cr.FlatChangePercent > criticalThreshold:
			return "Critical regression! Immediate investigation required."
		case cr.FlatChangePercent > significantThreshold:
			return "Significant regression detected. Consider rollback or optimization."
		case cr.FlatChangePercent > moderateThreshold:
			return "Moderate regression. Review recent changes and optimize if needed."
		default:
			return "Minor regression detected. Monitor for trends."
		}
	default:
		return "No action required. Continue monitoring."
	}
}

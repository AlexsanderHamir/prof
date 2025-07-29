package tracker

import (
	"time"
)

type ProfileChangeReport struct {
	FunctionChanges []*FunctionChangeResult
}

type FunctionChangeResult struct {
	FunctionName      string
	ChangeType        string // shared.REGRESSION, shred.IMPROVEMENT, shared.STABLE
	FlatChangePercent float64
	CumChangePercent  float64
	FlatAbsolute      struct {
		Before float64
		After  float64
		Delta  float64
	}
	CumAbsolute struct {
		Before float64
		After  float64
		Delta  float64
	}
	Timestamp time.Time
}

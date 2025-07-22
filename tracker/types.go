package tracker

import (
	"time"
)

type ProfileChangeReport struct {
	FunctionChanges []*FunctionChangeResult
}

type ProfileChangeSummary struct {
	TotalFunctions    int
	Regressions       int
	Improvements      int
	Stable            int
	NewFunctions      int
	DeletedFunctions  int
	WorstRegression   *FunctionChangeResult // Function with biggest regression
	BestImprovement   *FunctionChangeResult // Function with biggest improvement
	AverageFlatChange float64
}

type ChangeMetadata struct {
	Environment string
	Version     string
	TestType    string // "benchmark", "load-test", "production"
	Tags        map[string]string
}

type FunctionChangeResult struct {
	FunctionName      string
	ChangeType        string // "REGRESSION", "IMPROVEMENT", "STABLE"
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

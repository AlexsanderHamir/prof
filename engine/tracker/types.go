package tracker

import (
	"time"
)

type ProfileChangeReport struct {
	FunctionChanges []*FunctionChangeResult
}

type AbsoluteChange struct {
	Before float64 `json:"before"`
	After  float64 `json:"after"`
	Delta  float64 `json:"delta"`
}

type FunctionChangeResult struct {
	FunctionName      string         `json:"function_name"`
	ChangeType        string         `json:"change_type"`
	FlatChangePercent float64        `json:"flat_change_percent"`
	CumChangePercent  float64        `json:"cum_change_percent"`
	FlatAbsolute      AbsoluteChange `json:"flat_absolute"`
	CumAbsolute       AbsoluteChange `json:"cum_absolute"`
	Timestamp         time.Time      `json:"timestamp"`
}

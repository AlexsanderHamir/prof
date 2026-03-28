package tracker

import (
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

func TestApplyCommandLineThresholdsNoFail(t *testing.T) {
	r := &ProfileChangeReport{
		FunctionChanges: []*FunctionChangeResult{
			{FunctionName: "f", FlatChangePercent: 50, ChangeType: internal.REGRESSION},
		},
	}
	sel := &Selections{UseThreshold: false}
	if err := applyCommandLineThresholds(r, sel); err != nil {
		t.Fatal(err)
	}
}

func TestApplyCommandLineThresholdsTriggers(t *testing.T) {
	r := &ProfileChangeReport{
		FunctionChanges: []*FunctionChangeResult{
			{FunctionName: "f", FlatChangePercent: 99, ChangeType: internal.REGRESSION},
		},
	}
	sel := &Selections{UseThreshold: true, RegressionThreshold: 10}
	if err := applyCommandLineThresholds(r, sel); err == nil {
		t.Fatal("expected regression failure")
	}
}

func TestShouldIgnoreFunctionByConfig(t *testing.T) {
	cfg := &internal.CITrackingConfig{
		IgnoreFunctions: []string{"x"},
		IgnorePrefixes:  []string{"runtime."},
	}
	if !shouldIgnoreFunctionByConfig(cfg, "x") {
		t.Fatal()
	}
	if !shouldIgnoreFunctionByConfig(cfg, "runtime.gc") {
		t.Fatal()
	}
	if shouldIgnoreFunctionByConfig(cfg, "other") {
		t.Fatal()
	}
}

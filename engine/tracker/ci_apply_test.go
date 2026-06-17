package tracker

import (
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"
)

func TestApplyCommandLineThresholdsNoFail(t *testing.T) {
	r := &ProfileChangeReport{
		FunctionChanges: []*FunctionChangeResult{
			{FunctionName: "f", FlatChangePercent: 50, ChangeType: ChangeRegression},
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
			{FunctionName: "f", FlatChangePercent: 99, ChangeType: ChangeRegression},
		},
	}
	sel := &Selections{UseThreshold: true, RegressionThreshold: 10}
	if err := applyCommandLineThresholds(r, sel); err == nil {
		t.Fatal("expected regression failure")
	}
}

func TestShouldIgnoreFunctionByConfig(t *testing.T) {
	cfg := &config.CITrackingConfig{
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

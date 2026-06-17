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

func TestShouldIgnoreFunctionByTrackPolicy(t *testing.T) {
	policy := config.TrackPolicy{
		IgnoreFunctions: []string{"x"},
		IgnorePrefixes:  []string{"runtime."},
	}
	if !config.ShouldIgnoreFunction(policy, "x") {
		t.Fatal()
	}
	if !config.ShouldIgnoreFunction(policy, "runtime.gc") {
		t.Fatal()
	}
	if config.ShouldIgnoreFunction(policy, "other") {
		t.Fatal()
	}
}

func TestApplyTrackThresholdsOnlyTriggers(t *testing.T) {
	r := &ProfileChangeReport{
		FunctionChanges: []*FunctionChangeResult{
			{FunctionName: "f", FlatChangePercent: 20, ChangeType: ChangeRegression},
		},
	}
	policy := config.TrackPolicy{MaxRegressionPercent: 10}
	if err := applyTrackThresholdsOnly(r, policy); err == nil {
		t.Fatal("expected regression failure")
	}
}

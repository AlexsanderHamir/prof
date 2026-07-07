package datamap

import (
	"github.com/AlexsanderHamir/prof/internal/pprofscale"
	"github.com/AlexsanderHamir/prof/parser"
)

func profileOutputUnit(d *parser.ProfileData) string {
	if d == nil || d.SampleUnit == "" {
		return "auto"
	}
	return pprofscale.SelectOutputUnit(d.SampleUnit, d.Total, d.Flat, d.Cum)
}

func profileTotalDisplay(d *parser.ProfileData) (display string, seconds float64, outputUnit string) {
	if d == nil || d.SampleUnit == "" {
		return "", 0, ""
	}
	outputUnit = profileOutputUnit(d)
	display = pprofscale.ScaledLabel(d.Total, d.SampleUnit, outputUnit)
	if s, ok := pprofscale.Seconds(d.Total, d.SampleUnit); ok {
		seconds = pprofscale.RoundSeconds(s)
	}
	return display, seconds, outputUnit
}

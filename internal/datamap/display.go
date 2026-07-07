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

func sampleDisplays(flat, cum int64, unit, outputUnit string) (flatDisplay, cumDisplay string, flatSec, cumSec float64) {
	if unit == "" {
		return "", "", 0, 0
	}
	flatDisplay = pprofscale.ScaledLabel(flat, unit, outputUnit)
	cumDisplay = pprofscale.ScaledLabel(cum, unit, outputUnit)
	if s, ok := pprofscale.Seconds(flat, unit); ok {
		flatSec = pprofscale.RoundSeconds(s)
	}
	if s, ok := pprofscale.Seconds(cum, unit); ok {
		cumSec = pprofscale.RoundSeconds(s)
	}
	return flatDisplay, cumDisplay, flatSec, cumSec
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

// Package pprofscale formats profile sample values using the same unit rules as
// go tool pprof (github.com/google/pprof/internal/measurement). Logic is copied
// so prof can match pprof -top output without importing pprof internal packages.
package pprofscale

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
)

// Scale converts value from fromUnit to toUnit. Returns scaled value and target unit
// (empty when uninteresting). Matches pprof measurement.Scale.
func Scale(value int64, fromUnit, toUnit string) (float64, string) {
	if value < 0 && -value > 0 {
		v, u := Scale(-value, fromUnit, toUnit)
		return -v, u
	}
	for _, ut := range unitTypes {
		if v, u, ok := ut.convertUnit(value, fromUnit, toUnit); ok {
			return v, u
		}
	}
	switch toUnit {
	case "count", "sample", "unit", "minimum", "auto":
		return float64(value), ""
	default:
		return float64(value), toUnit
	}
}

// ScaledLabel formats value like pprof -top (two decimal places in outputUnit).
func ScaledLabel(value int64, fromUnit, toUnit string) string {
	v, u := Scale(value, fromUnit, toUnit)
	sv := strings.TrimSuffix(fmt.Sprintf("%.2f", v), ".00")
	if sv == "0" || sv == "-0" {
		return "0"
	}
	return sv + u
}

// Seconds converts a sample value to seconds when fromUnit is a time unit; ok is false otherwise.
func Seconds(value int64, fromUnit string) (sec float64, ok bool) {
	v, u := Scale(value, fromUnit, "s")
	if u != "s" {
		return 0, false
	}
	return v, true
}

// Unit includes aliases for a specific unit and factor to the base unit in its category.
type Unit struct {
	CanonicalName string
	aliases       []string
	Factor        float64
}

// UnitType groups units in one category (memory, time, …) with a default unit.
type UnitType struct {
	DefaultUnit Unit
	Units       []Unit
}

func (ut UnitType) findByAlias(alias string) *Unit {
	for _, u := range ut.Units {
		if slices.Contains(u.aliases, alias) {
			return &u
		}
	}
	return nil
}

func (ut UnitType) sniffUnit(unit string) *Unit {
	unit = strings.ToLower(unit)
	if len(unit) > 2 {
		unit = strings.TrimSuffix(unit, "s")
	}
	return ut.findByAlias(unit)
}

func (ut UnitType) autoScale(value float64) (float64, string, bool) {
	var f float64
	var unit string
	for _, u := range ut.Units {
		if u.Factor >= f && (value/u.Factor) >= 1.0 {
			f = u.Factor
			unit = u.CanonicalName
		}
	}
	if f == 0 {
		return 0, "", false
	}
	return value / f, unit, true
}

func (ut UnitType) convertUnit(value int64, fromUnitStr, toUnitStr string) (float64, string, bool) {
	fromUnit := ut.sniffUnit(fromUnitStr)
	if fromUnit == nil {
		return 0, "", false
	}
	v := float64(value) * fromUnit.Factor
	if toUnitStr == "minimum" || toUnitStr == "auto" {
		if v, u, ok := ut.autoScale(v); ok {
			return v, u, true
		}
		return v / ut.DefaultUnit.Factor, ut.DefaultUnit.CanonicalName, true
	}
	toUnit := ut.sniffUnit(toUnitStr)
	if toUnit == nil {
		return v / ut.DefaultUnit.Factor, ut.DefaultUnit.CanonicalName, true
	}
	return v / toUnit.Factor, toUnit.CanonicalName, true
}

// unitTypes matches github.com/google/pprof/internal/measurement.UnitTypes.
var unitTypes = []UnitType{{
	Units: []Unit{
		{"B", []string{"b", "byte"}, 1},
		{"kB", []string{"kb", "kbyte", "kilobyte"}, float64(1 << 10)},
		{"MB", []string{"mb", "mbyte", "megabyte"}, float64(1 << 20)},
		{"GB", []string{"gb", "gbyte", "gigabyte"}, float64(1 << 30)},
		{"TB", []string{"tb", "tbyte", "terabyte"}, float64(1 << 40)},
		{"PB", []string{"pb", "pbyte", "petabyte"}, float64(1 << 50)},
	},
	DefaultUnit: Unit{"B", []string{"b", "byte"}, 1},
}, {
	Units: []Unit{
		{"ns", []string{"ns", "nanosecond"}, float64(time.Nanosecond)},
		{"us", []string{"μs", "us", "microsecond"}, float64(time.Microsecond)},
		{"ms", []string{"ms", "millisecond"}, float64(time.Millisecond)},
		{"s", []string{"s", "sec", "second"}, float64(time.Second)},
		{"hrs", []string{"hour", "hr"}, float64(time.Hour)},
	},
	DefaultUnit: Unit{"s", []string{}, float64(time.Second)},
}}

// RoundSeconds rounds to two decimal places like pprof ScaledLabel before stripping .00.
func RoundSeconds(sec float64) float64 {
	return math.Round(sec*100) / 100
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// SelectOutputUnit picks the display unit go tool pprof -top uses for an entire report.
// Mirrors github.com/google/pprof/internal/report.Report.selectOutputUnit.
func SelectOutputUnit(sampleUnit string, total int64, flat, cum map[string]int64) string {
	if sampleUnit == "" {
		return "auto"
	}
	var minValue int64
	seen := make(map[string]struct{}, len(flat))
	for sym, f := range flat {
		seen[sym] = struct{}{}
		nodeMin := abs64(f)
		if nodeMin == 0 {
			nodeMin = abs64(cum[sym])
		}
		if nodeMin > 0 && (minValue == 0 || nodeMin < minValue) {
			minValue = nodeMin
		}
	}
	for sym, c := range cum {
		if _, ok := seen[sym]; ok {
			continue
		}
		nodeMin := abs64(c)
		if nodeMin > 0 && (minValue == 0 || nodeMin < minValue) {
			minValue = nodeMin
		}
	}
	maxValue := total
	if minValue == 0 {
		minValue = maxValue
	}
	_, minUnit := Scale(minValue, sampleUnit, "minimum")
	_, maxUnit := Scale(maxValue, sampleUnit, "minimum")
	unit := minUnit
	if minUnit != maxUnit && minValue*100 < maxValue {
		_, unit = Scale(100*minValue, sampleUnit, "minimum")
	}
	if unit != "" {
		return unit
	}
	return sampleUnit
}

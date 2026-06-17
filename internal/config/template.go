package config

import "strings"

const docSiteBase = "https://alexsanderhamir.github.io/prof"

// ExampleTemplate returns prof.json.example content with // comments explaining each field.
func ExampleTemplate(modulePath string) string {
	modulePath = strings.TrimSpace(modulePath)
	includeExample := modulePath
	if includeExample == "" {
		includeExample = "github.com/you/yourmodule"
	}

	return strings.TrimRight(`
// prof.json.example — reference for prof (https://github.com/AlexsanderHamir/prof)
// Generated prof.json is minimal ({ "version": 1 }); this file shows optional sections.
// Copy sections into prof.json as needed. This file is not loaded by prof; prof.json is valid JSON without comments.
// Full guide: `+docSiteBase+`/configure/
// File location: `+docSiteBase+`/workspace/#profjson

{
    "version": 1,

    // collection — which functions prof keeps when saving CPU/memory profiles after benchmarks.
    // Docs: `+docSiteBase+`/configure/#collection
    //       `+docSiteBase+`/collect/#artifact-layout-under-benchtag
    "collection": {
        "defaults": {
            // include_prefixes: if set, only functions whose full pprof symbol contains one of these
            // substrings (usually your module import path) are saved into per-function extracts and grouped reports.
            // Example: "`+includeExample+`" or "`+includeExample+`/internal/foo"
            // Empty [] keeps every function (including stdlib) — usually too broad.
            "include_prefixes": [
                "`+jsonString(includeExample)+`"
            ],

            // ignore_functions: skip these short function names even when include_prefixes matches.
            // Applied together: a function must match a prefix AND not appear in this list.
            "ignore_functions": [
                "init",
                "TestMain",
                "BenchmarkMain"
            ]
        },

        // Optional — override defaults for one benchmark (prof auto). Key = benchmark name:
        // Docs: `+docSiteBase+`/configure/#collection-benchmarks
        "benchmarks": {
            "BenchmarkMyHotPath": {
                "include_prefixes": ["`+includeExample+`/pkg/hot"],
                "ignore_functions": ["BenchmarkHelper"]
            }
        },

        // Optional — override for one manual profile file (prof manual). Key = file stem, e.g. BenchmarkFoo_cpu:
        // Docs: `+docSiteBase+`/configure/#collection-manual-profiles
        //       `+docSiteBase+`/collect/#prof-manual
        "manual_profiles": {
            "BenchmarkFoo_cpu": {
                "include_prefixes": ["`+includeExample+`/pkg/foo"]
            }
        }
    },

    // track — built-in regression check (prof track / UI compare): when to ignore noise and when to fail.
    // Docs: `+docSiteBase+`/configure/#track
    //       `+docSiteBase+`/compare/#regression-gate
    //       `+docSiteBase+`/ci/#json-in-profjson
    "track": {
        "defaults": {
            // ignore_prefixes: skip functions whose full name starts with one of these strings (runtime/test noise).
            "ignore_prefixes": [
                "runtime.",
                "reflect.",
                "testing."
            ],

            // ignore_functions: skip exact full symbol names from regression reports (optional).
            "ignore_functions": [],

            // min_change_percent: ignore regressions smaller than this percent (noise floor).
            "min_change_percent": 5.0,

            // max_regression_percent: fail when worst flat regression meets or exceeds this percent (0 = never fail).
            "max_regression_percent": 15.0,

            // fail_on_improvement: set true to fail on unexpected speedups too (unusual; default is false).
            "fail_on_improvement": false
        },

        // Optional — stricter limits for one benchmark:
        // Docs: `+docSiteBase+`/configure/#track-benchmarks
        //       `+docSiteBase+`/ci/#json-in-profjson
        "benchmarks": {
            "BenchmarkCritical": {
                "max_regression_percent": 5.0
            }
        }
    }
}
`, "\n") + "\n"
}

func jsonString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

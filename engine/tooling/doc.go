// Package tooling centralizes subprocess execution and command-line construction
// for external tools (for example go, go tool pprof, benchstat). It provides a
// small [Runner] boundary for tests and a typed catalog for profile kinds and
// pprof argument presets. In-process profile parsing remains in package parser.
package tooling

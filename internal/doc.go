// Package internal holds shared configuration ([Config], [BenchArgs]), path constants ([const.go]),
// config/template IO ([PrintConfiguration], [LoadFromFile]), and thin wrappers around lower-level helpers.
//
// Related packages:
//
//   - [github.com/AlexsanderHamir/prof/internal/repofs] — module root discovery and tag directory resets
//   - [github.com/AlexsanderHamir/prof/internal/testpaths] — test helpers for locating fixture assets under tests/assets
//
// Keep feature orchestration in engine packages and compose them via [app.Services] at the CLI boundary.
package internal

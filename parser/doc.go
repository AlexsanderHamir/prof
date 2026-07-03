// Package parser reads Go pprof profiles and builds flat/cumulative summaries and function lists.
//
// Structure:
//
//   - [Pipeline] and stage interfaces — swap open/decode/validate/aggregation without changing callers.
//   - Types in types.go — [ProfileData], report structs.
//   - profile_io.go — load/parse/validate entrypoints wired to the default pipeline.
//   - aggregate.go — sample → flat/cum maps and percentages.
//   - symbol_name.go — function string parsing for filters.
//   - facade.go — path-based API: [GetAllFunctionNames] and *V2 equivalents kept for stable call sites.
package parser

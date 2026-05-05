// Package parser reads Go pprof profiles and builds flat/cumulative summaries, function lists,
// and package-grouped markdown reports.
//
// Structure:
//
//   - [Pipeline] and stage interfaces — swap open/decode/validate/aggregation without changing callers.
//   - Types in types.go — [ProfileData], [LineObj], report structs.
//   - profile_io.go — load/parse/validate entrypoints wired to the default pipeline.
//   - aggregate.go — sample → flat/cum maps and percentages.
//   - symbol_name.go — function/package string parsing for filters and grouping.
//   - package_report.go — markdown formatting for package groups.
//   - facade.go — path-based API: [TurnLinesIntoObjects], [GetAllFunctionNames], [OrganizeProfileByPackage],
//     and *V2 equivalents kept for stable call sites.
package parser

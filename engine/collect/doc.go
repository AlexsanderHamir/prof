// Package collect runs auto and manual profile collection under .prof/<tag>/.
// Artifacts are grouped by data domain: profiles, measurements, hotspots,
// source_lines, and call_graphs (see internal/workspace.TagLayout).
// source_lines extraction fans out go tool pprof -list subprocesses with bounded parallelism.
package collect

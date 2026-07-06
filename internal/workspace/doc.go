// Package workspace defines .prof/<tag>/ layout paths and tag lifecycle helpers.
//
// Artifact domains under each tag describe the data they hold (domain/benchmark/artifact):
//
//   - profiles/      — raw pprof profile binaries (e.g. cpu.out)
//   - measurements/  — go test benchmark run stats (run.txt)
//   - hotspots/      — function-ranked stack summaries per profile (cpu.txt)
//   - source_lines/  — line-level pprof -list extracts per profile kind
//   - call_graphs/   — optional Graphviz PNG call graphs per profile kind
//   - notes.txt      — tag-level note at the tag root
package workspace

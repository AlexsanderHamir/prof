// Package benchmark runs `go test` benchmarks with profiling flags, materializes outputs under
// bench/<tag>/, drives pprof text/PNG and per-function extraction via [collector], and discovers
// Benchmark* functions in the module.
package benchmark

// Package tooling centralizes subprocess execution and command-line construction
// for external tools (for example go, go tool pprof, benchstat). Production code
// must spawn processes through [ExecRunner], [StartDetached], or [LookPath] here
// so golangci-lint forbidigo can block raw [exec.Command] outside tests.
// In-process profile parsing remains in package parser.
package tooling

// Package cursoragent runs the Cursor CLI tool "cursor-agent" in non-interactive
// mode (--print --output-format stream-json) for use by higher-level prof features.
//
// Prerequisites:
//   - Install Cursor and ensure "cursor-agent" is on PATH, or set [EnvBinaryOverride]
//     to the full path to the agent binary, or pass [Options.BinaryPath] when constructing
//     a [Client] (callers typically map the PROF_CURSOR_AGENT environment variable and
//     a future --cursor-agent flag into Options.BinaryPath; precedence is flag > env > default name).
//
// This package does not invoke the binary until [Client.Probe] or [Client.Run] is called.
// It does not persist prompts or results to disk.
//
// Adapted from patterns in github.com/AlexsanderHamir/T2A (pkgs/agents/runner/cursor and adapterkit);
// prof does not import T2A as a module.
package cursoragent

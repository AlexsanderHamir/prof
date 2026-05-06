package cursoragent

// FixBinaryHelpBlock returns operator guidance when cursor-agent cannot be found or run.
// Keep in sync with package [doc] and future Cobra command Long text.
func FixBinaryHelpBlock() string {
	return `cursor-agent is required but was not found or failed to run.

Fix one of:
  1) Add "cursor-agent" to your PATH (install the Cursor CLI / agent shim), or
  2) Set environment variable PROF_CURSOR_AGENT to the full path of the agent executable, or
  3) Pass --cursor-agent <path> on commands that support it (when wired in the CLI).

This must be the "cursor-agent" tool, not the "cursor" editor launcher.
See: https://cursor.com/docs`
}

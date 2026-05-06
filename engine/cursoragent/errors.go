package cursoragent

import "errors"

// Typed errors for [Client.Run] and [Client.Probe]. Use [errors.Is] for classification.
var (
	// ErrTimeout indicates the run exceeded its context deadline or configured timeout.
	ErrTimeout = errors.New("cursoragent: timeout")

	// ErrNonZeroExit indicates cursor-agent exited with a non-zero status or reported is_error.
	ErrNonZeroExit = errors.New("cursoragent: non-zero exit")

	// ErrInvalidOutput indicates stdout could not be parsed as usable stream-json output.
	ErrInvalidOutput = errors.New("cursoragent: invalid output")

	// ErrBinaryNotFound indicates the configured binary path could not be resolved or does not exist.
	ErrBinaryNotFound = errors.New("cursoragent: binary not found")
)

package tooling

import "os/exec"

// LookPath resolves an executable name on PATH. Callers use this instead of importing [os/exec] only for [exec.LookPath].
func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

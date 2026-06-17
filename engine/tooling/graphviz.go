package tooling

import "os/exec"

// lookPathFn is overridden in tests via LookPathForTests.
var lookPathFn = exec.LookPath

// LookPathForTests overrides PATH lookup when set (tests only).
var LookPathForTests func(string) (string, error)

func pathLook(name string) (string, error) {
	if LookPathForTests != nil {
		return LookPathForTests(name)
	}
	return lookPathFn(name)
}

// GraphvizAvailable reports whether the Graphviz `dot` binary is on PATH.
func GraphvizAvailable() bool {
	_, err := pathLook("dot")
	return err == nil
}

// SkipPNGNotice is printed when PNG generation is auto-skipped.
const SkipPNGNotice = "Graphviz not found; skipping PNG generation (text profiles still collected)"

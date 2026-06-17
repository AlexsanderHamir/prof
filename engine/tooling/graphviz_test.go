package tooling

import "testing"

func TestGraphvizAvailable(t *testing.T) {
	orig := LookPathForTests
	t.Cleanup(func() { LookPathForTests = orig })

	LookPathForTests = func(name string) (string, error) {
		if name == "dot" {
			return "/usr/bin/dot", nil
		}
		return "", execErr("not found")
	}
	if !GraphvizAvailable() {
		t.Fatal("expected available when dot is on PATH")
	}

	LookPathForTests = func(string) (string, error) {
		return "", execErr("not found")
	}
	if GraphvizAvailable() {
		t.Fatal("expected unavailable when dot is missing")
	}
}

type execErr string

func (e execErr) Error() string { return string(e) }

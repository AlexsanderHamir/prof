package collect

import (
	"testing"

	"github.com/AlexsanderHamir/prof/engine/tooling"
)

func TestApplyAutoSkipPNG(t *testing.T) {
	orig := tooling.LookPathForTests
	t.Cleanup(func() { tooling.LookPathForTests = orig })

	tooling.LookPathForTests = func(name string) (string, error) {
		if name == "dot" {
			return "/bin/dot", nil
		}
		return "", errNotFound("missing")
	}
	opts := AutoOptions{SkipPNG: false}
	if applyAutoSkipPNG(&opts) {
		t.Fatal("expected no change when graphviz available")
	}
	if opts.SkipPNG {
		t.Fatal("SkipPNG should stay false")
	}

	tooling.LookPathForTests = func(string) (string, error) {
		return "", errNotFound("missing")
	}
	opts = AutoOptions{SkipPNG: false}
	if !applyAutoSkipPNG(&opts) {
		t.Fatal("expected auto skip when graphviz missing")
	}
	if !opts.SkipPNG {
		t.Fatal("SkipPNG should be true")
	}

	opts = AutoOptions{SkipPNG: true}
	if applyAutoSkipPNG(&opts) {
		t.Fatal("expected no change when already skipping")
	}
}

type errNotFound string

func (e errNotFound) Error() string { return string(e) }

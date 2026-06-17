package cli

import (
	"errors"
	"testing"
)

func TestErrUILoopExit(t *testing.T) {
	t.Parallel()
	if !errors.Is(errUILoopExit, errUILoopExit) {
		t.Fatal("errUILoopExit should match itself")
	}
}

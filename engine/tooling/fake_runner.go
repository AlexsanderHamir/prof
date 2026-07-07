package tooling

import (
	"context"
	"errors"
	"sync"
)

// FakeRun records one invocation of [FakeRunner.Run].
type FakeRun struct {
	Argv []string
	Opts RunOpts
}

// FakeRunner implements [Runner] for tests. It records each run and returns configured results in order.
type FakeRunner struct {
	mu   sync.Mutex
	Runs []FakeRun
	Out  [][]byte
	Err  []error
	n    int
}

// Run appends argv and opts to Runs, then returns the next configured Out/Err pair.
func (f *FakeRunner) Run(ctx context.Context, argv []string, opts RunOpts) ([]byte, error) {
	_ = ctx
	if f == nil {
		return nil, errors.New("tooling: nil FakeRunner")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Runs = append(f.Runs, FakeRun{Argv: append([]string(nil), argv...), Opts: opts})
	i := f.n
	f.n++
	var out []byte
	var err error
	if i < len(f.Out) {
		out = f.Out[i]
	}
	if i < len(f.Err) {
		err = f.Err[i]
	} else if i >= len(f.Out) && len(f.Err) == 0 {
		err = errors.New("tooling: FakeRunner missing Out/Err entry")
	}
	return out, err
}

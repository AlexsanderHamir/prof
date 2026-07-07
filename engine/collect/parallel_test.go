package collect

import (
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
)

func TestParallelFor_allIndicesRun(t *testing.T) {
	t.Parallel()
	const n = 20
	var seen atomic.Int64
	errs := parallelFor(n, 4, func(i int) error {
		seen.Add(1 << i)
		return nil
	})
	if len(errs) != n {
		t.Fatalf("len(errs)=%d want %d", len(errs), n)
	}
	var mask int64
	for i := 0; i < n; i++ {
		mask |= 1 << i
	}
	if seen.Load() != mask {
		t.Fatalf("not all indices executed: seen=%b want=%b", seen.Load(), mask)
	}
}

func TestParallelFor_preservesErrorsByIndex(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("boom")
	errs := parallelFor(8, 4, func(i int) error {
		if i == 2 || i == 5 {
			return wantErr
		}
		return nil
	})
	for i, err := range errs {
		if i == 2 || i == 5 {
			if !errors.Is(err, wantErr) {
				t.Fatalf("index %d: got %v want %v", i, err, wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("index %d: unexpected error %v", i, err)
		}
	}
}

func TestSourceLinesWorkers_capsAtDefault(t *testing.T) {
	t.Parallel()
	if runtime.GOMAXPROCS(0) < defaultSourceLinesWorkers {
		t.Skip("GOMAXPROCS below default cap")
	}
	if got := sourceLinesWorkers(100); got != defaultSourceLinesWorkers {
		t.Fatalf("sourceLinesWorkers(100)=%d want %d", got, defaultSourceLinesWorkers)
	}
}

func TestSourceLinesWorkers_respectsJobCount(t *testing.T) {
	t.Parallel()
	if got := sourceLinesWorkers(3); got != 3 {
		t.Fatalf("sourceLinesWorkers(3)=%d want 3", got)
	}
}

func TestSourceLinesWorkers_zeroJobs(t *testing.T) {
	t.Parallel()
	if got := sourceLinesWorkers(0); got != 0 {
		t.Fatalf("sourceLinesWorkers(0)=%d want 0", got)
	}
}

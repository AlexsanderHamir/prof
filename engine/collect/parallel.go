package collect

import (
	"runtime"
	"sync"
)

// defaultSourceLinesWorkers caps concurrent go tool pprof -list subprocesses.
// Each worker loads the profile binary into memory.
const defaultSourceLinesWorkers = 8

// sourceLinesWorkers returns the worker count for source_lines subprocess fan-out.
func sourceLinesWorkers(jobCount int) int {
	if jobCount <= 0 {
		return 0
	}
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	if jobCount < workers {
		workers = jobCount
	}
	if workers > defaultSourceLinesWorkers {
		workers = defaultSourceLinesWorkers
	}
	return workers
}

// parallelFor runs fn(i) for i in [0,n) with at most workers goroutines.
// It returns per-index errors; nil entries mean success. All jobs run to completion.
func parallelFor(n, workers int, fn func(i int) error) []error {
	if n == 0 {
		return nil
	}
	errs := make([]error, n)
	if n == 1 || workers <= 1 {
		for i := range n {
			errs[i] = fn(i)
		}
		return errs
	}
	if workers > n {
		workers = n
	}
	jobs := make(chan int, n)
	for i := range n {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for i := range jobs {
				errs[i] = fn(i)
			}
		}()
	}
	wg.Wait()
	return errs
}

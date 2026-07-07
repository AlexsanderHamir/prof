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
	max := runtime.GOMAXPROCS(0)
	if max < 1 {
		max = 1
	}
	if jobCount < max {
		max = jobCount
	}
	if max > defaultSourceLinesWorkers {
		max = defaultSourceLinesWorkers
	}
	return max
}

// parallelFor runs fn(i) for i in [0,n) with at most workers goroutines.
// It returns per-index errors; nil entries mean success. All jobs run to completion.
func parallelFor(n, workers int, fn func(i int) error) []error {
	if n == 0 {
		return nil
	}
	errs := make([]error, n)
	if n == 1 || workers <= 1 {
		for i := 0; i < n; i++ {
			errs[i] = fn(i)
		}
		return errs
	}
	if workers > n {
		workers = n
	}
	jobs := make(chan int, n)
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
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

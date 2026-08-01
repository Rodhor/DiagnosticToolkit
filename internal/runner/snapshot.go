package runner

import (
	"context"
	"runtime"
	"sync"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

func (r *Runner) RunOnce(ctx context.Context, out chan<- domain.CheckResult) []domain.CheckResult {
	if out != nil {
		defer close(out)
	}
	if len(r.checks) == 0 {
		return nil
	}

	// Define needed variables for concurrent execution
	var (
		mu          sync.Mutex
		accumulated []domain.CheckResult
	)
	processAndAccumulate := func(res domain.CheckResult, rc RegisteredCheck) bool {
		processedRes, ok := r.processOne(ctx, out, res, rc)
		if !ok {
			return false
		}
		mu.Lock()
		accumulated = append(accumulated, processedRes)
		mu.Unlock()
		return true
	}

	// 1. Partition checks into Serial and Parallel sets (Serial: checks that run sequentially, Parallel: checks that run concurrently)
	var serialChecks, parallelChecks []RegisteredCheck
	for _, rc := range r.checks {
		if rc.Serial {
			serialChecks = append(serialChecks, rc)
		} else {
			parallelChecks = append(parallelChecks, rc)
		}
	}

	// 2. Execute Serial checks sequentially
OuterSerial:
	for _, rc := range serialChecks {
		if ctx.Err() != nil {
			break
		}

		results := rc.Check.Run(ctx)
		for _, res := range results {
			if !processAndAccumulate(res, rc) {
				break OuterSerial
			}
		}
	}

	// 3. Execute Parallel checks
	if len(parallelChecks) > 0 && ctx.Err() == nil {
		concurrency := r.opts.Concurrency
		if concurrency <= 0 {
			concurrency = runtime.GOMAXPROCS(0)
		}

		jobs := make(chan RegisteredCheck, len(parallelChecks))
		for _, rc := range parallelChecks {
			jobs <- rc
		}

		close(jobs)

		var wg sync.WaitGroup
		workerCount := min(concurrency, len(parallelChecks))
		wg.Add(workerCount)

		for range workerCount {
			go func() {
				defer wg.Done()
				for rc := range jobs {
					if ctx.Err() != nil {
						return
					}

					results := rc.Check.Run(ctx)
					for _, result := range results {
						if !processAndAccumulate(result, rc) {
							return
						}
					}
				}
			}()
		}
		wg.Wait()
	}

	sortResults(accumulated)
	return accumulated

}

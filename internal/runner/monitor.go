package runner

import (
	"context"
	"sync"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

func (r *Runner) Monitor(ctx context.Context, out chan<- domain.CheckResult) {
	if out != nil {
		defer close(out)
	}
	if len(r.checks) == 0 {
		return
	}

	var wg sync.WaitGroup

	for _, rc := range r.checks {
		interval := rc.Interval
		if interval <= 0 {
			interval = r.opts.Interval
		}

		wg.Add(1)
		go func(rc RegisteredCheck, interval time.Duration) {
			defer wg.Done()

			// 1. Immediate initial execution

			if !r.runAndEmit(ctx, rc, out, interval) {
				return
			}

			// 2. Periodic execution via ticker
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				// if context is done, exit
				case <-ctx.Done():
					return

					// if ticker fires, run the check again
				case <-ticker.C:
					if !r.runAndEmit(ctx, rc, out, interval) {
						return
					}
				}
			}
		}(rc, interval)
	}

	wg.Wait()
}

func (r *Runner) runAndEmit(
	ctx context.Context,
	rc RegisteredCheck,
	out chan<- domain.CheckResult,
	interval time.Duration,
) bool {
	if ctx.Err() != nil {
		return false
	}

	start := time.Now()
	results := rc.Check.Run(ctx)
	duration := time.Since(start)

	for _, res := range results {
		if _, ok := r.processOne(ctx, out, res, rc); !ok {
			return false
		}
	}

	// Overrun detection - if the check took longer than the interval, emit a warning
	if interval > 0 && duration > interval {
		overrunRes := domain.NewResultBuilder(rc.CheckType, "self").
			Status(domain.Warn).
			Reason(domain.ReasonOverrun).
			Timestamp(start).
			Metrics([]domain.Metric{
				domain.DurationMetric(domain.KeyRuntimeMs, "Runtime", duration, true),
				domain.DurationMetric(domain.KeyIntervalMs, "Interval", interval, false),
			}).Build()
		if _, ok := r.processOne(ctx, out, overrunRes, rc); !ok {
			return false
		}
	}

	return true

}

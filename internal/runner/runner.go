package runner

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type RegisteredCheck struct {
	CheckType domain.CheckType
	Check     check.Check
	Name      string
	Category  domain.Category
	Interval  time.Duration
	Serial    bool
}

type Options struct {
	Thresholds  domain.Thresholds
	Concurrency int
	Interval    time.Duration
}

type Runner struct {
	checks []RegisteredCheck
	opts   Options
}

func New(checks []RegisteredCheck, opts Options) *Runner {
	return &Runner{
		checks: checks,
		opts:   opts,
	}
}

func (r *Runner) processOne(ctx context.Context, out chan<- domain.CheckResult, res domain.CheckResult, rc RegisteredCheck) (domain.CheckResult, bool) {
	res = stamp(res, rc)
	res = domain.ApplyThreshold(res, r.opts.Thresholds)

	ok := emit(ctx, out, res)
	return res, ok

}

func emit(ctx context.Context, out chan<- domain.CheckResult, result domain.CheckResult) bool {
	if out == nil {
		return true
	}
	select {
	case out <- result:
		return true
	case <-ctx.Done():
		return false
	}
}

func stamp(result domain.CheckResult, rc RegisteredCheck) domain.CheckResult {
	if rc.CheckType != "" {
		result.CheckType = rc.CheckType
	}
	result.CheckName = rc.Name
	result.Category = rc.Category
	return result
}

// sortResults orders check results deterministically by Category -> CheckName -> Target.
// Note: sortResults is ONLY invoked by RunOnce (snapshot mode) to provide deterministic static output.
// Monitor mode bypasses sorting because results are streamed live in real-time as they complete.
func sortResults(results []domain.CheckResult) {
	slices.SortFunc(results, func(a domain.CheckResult, b domain.CheckResult) int {
		if c := cmp.Compare(a.Category, b.Category); c != 0 {
			return c
		}
		if c := cmp.Compare(a.CheckName, b.CheckName); c != 0 {
			return c
		}
		return cmp.Compare(a.Target, b.Target)
	})
}

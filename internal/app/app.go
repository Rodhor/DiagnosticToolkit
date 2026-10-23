package app

import (
	"context"
	"sync"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/platform"
	"github.com/rodhor/DiagnosticToolset/internal/runner"
)

type App struct {
	cfg      Config
	provider platform.Provider

	mu       sync.RWMutex
	snapshot platform.Snapshot
	snapErr  error
	runner   *runner.Runner
}

func New(ctx context.Context, cfg Config) (*App, error) {
	a := &App{cfg: cfg, provider: platform.New()}
	a.Refresh(ctx)
	return a, nil
}

func (a *App) Refresh(ctx context.Context) error {
	snap, err := a.provider.Snapshot(ctx)

	var d derived
	if err == nil {
		d.Gateway = gatewayTarget(snap)
		d.InternalResolvers = discoveredResolvers(a.cfg.Resolvers, snap)
	}

	r := runner.New(registerChecks(a.cfg, d), runner.Options{
		Thresholds:  a.cfg.Thresholds,
		Concurrency: a.cfg.Concurrency,
		Interval:    a.cfg.Interval,
	})

	a.mu.Lock()
	a.snapshot, a.snapErr, a.runner = snap, err, r
	a.mu.Unlock()
	return err
}

func (a *App) Environment() (platform.Snapshot, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.snapshot, a.snapErr
}

func (a *App) RunOnce(ctx context.Context, out chan<- domain.CheckResult) []domain.CheckResult {
	return a.currentRunner().RunOnce(ctx, out)
}

func (a *App) Monitor(ctx context.Context, out chan<- domain.CheckResult) {
	a.currentRunner().Monitor(ctx, out)
}

func (a *App) currentRunner() *runner.Runner {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.runner
}

package server

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/report"
	"github.com/rodhor/DiagnosticToolset/web/templates/components"
)

type State struct {
	mu       sync.Mutex
	app      *app.App
	running  bool
	cancel   context.CancelFunc
	buffer   *ringBuffer[domain.CheckResult]
	hub      *sseHub
	summary  *report.Summary
	sumStart time.Time
}

func newState(a *app.App) *State {
	return &State{
		app:    a,
		buffer: newRingBuffer[domain.CheckResult](500),
		hub:    newSSEHub(),
	}
}

func (s *State) Start(ctx context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	runCtx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true
	s.summary = report.NewSummary(time.Now())
	s.sumStart = time.Now()
	out := make(chan domain.CheckResult, 64)
	go func() {
		s.app.Monitor(runCtx, out)
		s.mu.Lock()
		s.running = false
		s.cancel = nil
		s.mu.Unlock()
	}()
	go func() {
		for res := range out {
			s.buffer.Add(res)
			if s.summary != nil {
				s.summary.Add(res)
			}
			if b, err := s.renderResults(); err == nil {
				s.hub.Broadcast(b)
			}
		}
	}()
	return true
}

func (s *State) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	if s.summary != nil {
		s.summary.Finalize(time.Now())
	}
}

func (s *State) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *State) RunOnce(ctx context.Context) []domain.CheckResult {
	res := s.app.RunOnce(context.Background(), nil)
	for _, r := range res {
		s.buffer.Add(r)
	}
	return res
}

func (s *State) Recent() []domain.CheckResult { return s.buffer.All() }
func (s *State) Summary() *report.Summary     { return s.summary }
func (s *State) HasSummary() bool             { return s.summary != nil }

func (s *State) renderResults() ([]byte, error) {
	var buf bytes.Buffer
	if err := components.Results(s.buffer.All()).Render(context.Background(), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderResults exposes the rendered results table (used by SSE initial send).
func (s *State) RenderResults() ([]byte, error) { return s.renderResults() }
func (s *State) Hub() *sseHub                { return s.hub }
func (s *State) App() *app.App               { return s.app }

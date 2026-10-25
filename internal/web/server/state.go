package server

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type State struct {
	mu      sync.Mutex
	app     *app.App
	running bool
	cancel  context.CancelFunc
	buffer  *ringBuffer[domain.CheckResult]
	hub     *sseHub
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
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true
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
			if b, err := json.Marshal(res); err == nil {
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
}

func (s *State) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *State) RunOnce(ctx context.Context) []domain.CheckResult {
	res := s.app.RunOnce(ctx, nil)
	for _, r := range res {
		s.buffer.Add(r)
	}
	return res
}

func (s *State) Recent() []domain.CheckResult { return s.buffer.All() }
func (s *State) Hub() *sseHub                { return s.hub }
func (s *State) App() *app.App               { return s.app }

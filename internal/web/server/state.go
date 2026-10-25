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

type Run struct {
	ID       int
	Started  time.Time
	Ended    time.Time
	Running  bool
	Buffer   *ringBuffer[domain.CheckResult]
	Summary  *report.Summary
	cancel   context.CancelFunc
	done     chan struct{}
}

type State struct {
	mu          sync.Mutex
	app         *app.App
	configPath  string
	runs        []*Run
	current     *Run
	nextID      int
	hub         *sseHub
}

func newState(a *app.App, configPath string) *State {
	return &State{app: a, configPath: configPath, hub: newSSEHub()}
}

func (s *State) Start(ctx context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil && s.current.Running {
		return false
	}
	runCtx, cancel := context.WithCancel(context.Background())
	r := &Run{
		ID:      s.nextID + 1,
		Started: time.Now(),
		Running: true,
		Buffer:  newRingBuffer[domain.CheckResult](5000),
		Summary: report.NewSummary(time.Now()),
		cancel:  cancel,
		done:    make(chan struct{}),
	}
	s.nextID = r.ID
	s.runs = append(s.runs, r)
	s.current = r

	out := make(chan domain.CheckResult, 64)
	go func() {
		s.app.Monitor(runCtx, out)
		s.mu.Lock()
		r.Running = false
		r.Ended = time.Now()
		if r.Summary != nil {
			r.Summary.Finalize(r.Ended)
		}
		if s.current == r {
			s.current = nil
		}
		s.mu.Unlock()
		close(r.done)
	}()
	go func() {
		for res := range out {
			r.Buffer.Add(res)
			if r.Summary != nil {
				r.Summary.Add(res)
			}
			if b, err := s.renderResultsFor(r); err == nil {
				s.hub.Broadcast(b)
			}
		}
	}()
	_ = ctx
	return true
}

func (s *State) Stop() {
	s.mu.Lock()
	var done chan struct{}
	if s.current != nil {
		if s.current.cancel != nil {
			s.current.cancel()
		}
		done = s.current.done
	}
	s.mu.Unlock()
	if done != nil {
		select {
		case <-done:
		case <-time.After(15 * time.Second):
		}
	}
}

func (s *State) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current != nil && s.current.Running
}

func (s *State) RunOnce(ctx context.Context) []domain.CheckResult {
	res := s.app.RunOnce(context.Background(), nil)
	// store as a run too? keep as current-ish? For simplicity, add to a run if none running? Skip.
	return res
}

func (s *State) Runs() []*Run {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Run, len(s.runs))
	copy(out, s.runs)
	return out
}

func (s *State) Current() *Run {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

func (s *State) RunByID(id int) *Run {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.runs {
		if r.ID == id {
			return r
		}
	}
	return nil
}

func (s *State) Recent() []domain.CheckResult {
	if r := s.Current(); r != nil {
		return r.Buffer.All()
	}
	return nil
}

func (s *State) Summary() *report.Summary {
	if r := s.Current(); r != nil {
		return r.Summary
	}
	if len(s.runs) > 0 {
		return s.runs[len(s.runs)-1].Summary
	}
	return nil
}

func (s *State) HasSummary() bool { return s.Summary() != nil }

func (s *State) renderResultsFor(r *Run) ([]byte, error) {
	var buf bytes.Buffer
	if err := components.Results(r.Buffer.All()).Render(context.Background(), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *State) RenderResults() ([]byte, error) {
	if r := s.Current(); r != nil {
		return s.renderResultsFor(r)
	}
	if len(s.runs) > 0 {
		return s.renderResultsFor(s.runs[len(s.runs)-1])
	}
	var buf bytes.Buffer
	if err := components.Results(nil).Render(context.Background(), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *State) App() *app.App          { return s.app }
func (s *State) ConfigPath() string     { return s.configPath }
func (s *State) SetApp(a *app.App) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.app = a
}
func (s *State) Hub() *sseHub  { return s.hub }

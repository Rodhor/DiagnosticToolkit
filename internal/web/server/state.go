package server

import (
	"context"
	"sync"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/report"
)

type Run struct {
	ID      int
	Mode    string
	Started time.Time
	Ended   time.Time
	Running bool
	Buffer  *ringBuffer[domain.CheckResult]
	Summary *report.Summary
	cancel  context.CancelFunc
	done    chan struct{}
}

type State struct {
	mu         sync.Mutex
	app        *app.App
	configPath string
	runs       []*Run
	current    *Run
	nextID     int
}

func newState(a *app.App, configPath string) *State {
	return &State{app: a, configPath: configPath}
}

func (s *State) Start(_ context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil && s.current.Running {
		return false
	}
	runCtx, cancel := context.WithCancel(context.Background())
	r := &Run{
		ID:      s.nextID + 1,
		Mode:    "Monitoring",
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

	a := s.app
	out := make(chan domain.CheckResult, 64)
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		for res := range out {
			r.Buffer.Add(res)
			r.Summary.Add(res)
		}
	}()
	go func() {
		a.Monitor(runCtx, out)
		<-consumerDone
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

func (s *State) RunOnce(_ context.Context) []domain.CheckResult {
	s.mu.Lock()
	s.nextID++
	r := &Run{
		ID:      s.nextID,
		Mode:    "Run once",
		Started: time.Now(),
		Running: true,
		Buffer:  newRingBuffer[domain.CheckResult](5000),
		Summary: report.NewSummary(time.Now()),
		done:    make(chan struct{}),
	}
	s.runs = append(s.runs, r)
	a := s.app
	s.mu.Unlock()

	res := a.RunOnce(context.Background(), nil)
	for _, item := range res {
		r.Buffer.Add(item)
		r.Summary.Add(item)
	}
	s.mu.Lock()
	r.Running = false
	r.Ended = time.Now()
	r.Summary.Finalize(r.Ended)
	s.mu.Unlock()
	close(r.done)
	return res
}

func (s *State) Runs() []*Run {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Run, len(s.runs))
	for i, run := range s.runs {
		snapshot := *run
		out[i] = &snapshot
	}
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
			snapshot := *r
			return &snapshot
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

func (s *State) App() *app.App      { return s.app }
func (s *State) ConfigPath() string { return s.configPath }
func (s *State) SetApp(a *app.App) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.app = a
}

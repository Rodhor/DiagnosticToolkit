package report

import (
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

const maxSamples = 20_000

type Window struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Count int       `json:"count"`
}

type CheckSummary struct {
	Category       domain.Category `json:"category"`
	CheckName      string          `json:"check_name"`
	Target         string          `json:"target"`
	Total          int             `json:"total"`
	Passed         int             `json:"passed"`
	Warned         int             `json:"warned"`
	Failed         int             `json:"failed"`
	Unknown        int             `json:"unknown"`
	Failures       []Window        `json:"failures,omitempty"`
	LatencySamples int             `json:"latency_samples"`
	LatencyP50     time.Duration   `json:"-"`
	LatencyP95     time.Duration   `json:"-"`
	MaxLatency     time.Duration   `json:"-"`

	// Internal state tracking
	activeWindow *Window
	samples      []time.Duration
	sampleIdx    int
}

func (cs CheckSummary) MarshalJSON() ([]byte, error) {
	type Alias CheckSummary

	return json.Marshal(&struct {
		Alias
		LatencyP50Ms int64 `json:"latency_p50_ms"`
		LatencyP95Ms int64 `json:"latency_p95_ms"`
		MaxLatencyMs int64 `json:"max_latency_ms"`
	}{
		Alias:        Alias(cs),
		LatencyP50Ms: int64(cs.LatencyP50.Milliseconds()),
		LatencyP95Ms: int64(cs.LatencyP95.Milliseconds()),
		MaxLatencyMs: int64(cs.MaxLatency.Milliseconds()),
	})

}

type Summary struct {
	mu      sync.Mutex
	Started time.Time                `json:"started"`
	Ended   time.Time                `json:"ended"`
	Checks  map[string]*CheckSummary `json:"checks"`
}

func NewSummary(started time.Time) *Summary {
	return &Summary{
		Started: started,
		Checks:  make(map[string]*CheckSummary),
	}
}

func (s *Summary) Add(r domain.CheckResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := r.Key()
	cs, exists := s.Checks[key]
	if !exists {
		cs = &CheckSummary{
			Category:  r.Category,
			CheckName: r.CheckName,
			Target:    r.Target,
			samples:   []time.Duration{},
		}
		s.Checks[key] = cs
	}

	cs.Total++

	// 1. Increment counters based on result status
	switch r.Status {
	case domain.Pass:
		cs.Passed++
	case domain.Warn:
		cs.Warned++
	case domain.Fail:
		cs.Failed++
	case domain.Unknown:
		cs.Unknown++
	}

	// 2. Track contigous failure windows
	ts := r.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	if r.Status == domain.Fail {
		if cs.activeWindow == nil {
			cs.activeWindow = &Window{
				Start: ts,
				End:   ts,
				Count: 1,
			}
		} else {
			cs.activeWindow.End = ts
			cs.activeWindow.Count++
		}
	} else {
		if cs.activeWindow != nil {
			cs.Failures = append(cs.Failures, *cs.activeWindow)
			cs.activeWindow = nil
		}
	}

	// 3. Ring buffer for latencies
	if r.Status == domain.Pass || r.Status == domain.Warn {
		cs.LatencySamples++
		if len(cs.samples) < maxSamples {
			cs.samples = append(cs.samples, r.Latency)
		} else {
			cs.samples[cs.sampleIdx] = r.Latency
			cs.sampleIdx = (cs.sampleIdx + 1) % maxSamples
		}
	}
}

func (s *Summary) Finalize(ended time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Ended = ended
	for _, cs := range s.Checks {
		// Close any open windows if run ended during a failure state
		if cs.activeWindow != nil {
			cs.Failures = append(cs.Failures, *cs.activeWindow)
			cs.activeWindow = nil
		}

		if len(cs.samples) == 0 {
			continue
		}

		sorted := slices.Clone(cs.samples)
		slices.Sort(sorted)

		n := len(sorted)
		cs.MaxLatency = sorted[n-1]
		cs.LatencyP50 = sorted[pRank(n, 0.50)]
		cs.LatencyP95 = sorted[pRank(n, 0.95)]
	}
}

func pRank(n int, percentage float64) int {
	idx := int(float64(n) * percentage)
	if idx >= n {
		return n - 1
	}
	return idx
}

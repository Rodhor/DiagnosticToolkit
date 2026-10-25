package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	Expected       string          `json:"expected,omitempty"`
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
			Expected:  r.Expected,
			samples:   []time.Duration{},
		}
		s.Checks[key] = cs
	}

	cs.Total++
	if r.Expected != "" {
		cs.Expected = r.Expected
	}

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

func WriteSummary(w io.Writer, s *Summary) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var b strings.Builder

	// 1. Header (Time Window & Duration)
	duration := s.Ended.Sub(s.Started).Round(time.Second)
	writeSection(&b, "Diagnostic Run Summary")
	fmt.Fprintf(&b, "Time Window: %s → %s (%v)\n",
		s.Started.Format("2006-01-02 15:04:05"),
		s.Ended.Format("2006-01-02 15:04:05"),
		duration,
	)

	totalChecks := len(s.Checks)
	if totalChecks == 0 {
		b.WriteString("No checks were executed.\n")
		_, err := fmt.Fprint(w, b.String())
		return err
	}

	// 2. Aggregate Tallies
	checkList := make([]*CheckSummary, 0, totalChecks)
	var totalFailures, totalWarnings, totalUnknowns, totalPassed, totalObs int

	for _, cs := range s.Checks {
		checkList = append(checkList, cs)
		totalFailures += cs.Failed
		totalWarnings += cs.Warned
		totalUnknowns += cs.Unknown
		totalPassed += cs.Passed
		totalObs += cs.Total
	}

	// 3. Top-Level Status Tally
	if totalFailures == 0 && totalWarnings == 0 && totalUnknowns == 0 {
		fmt.Fprintf(&b, "STATUS: ALL PASSED (%d checks, 100%% successful across %d observations)\n\n",
			totalChecks, totalObs)
	} else {
		fmt.Fprintf(&b, "STATUS: %d failures, %d warnings, %d unknown across %d checks (%d total observations)\n\n",
			totalFailures, totalWarnings, totalUnknowns, totalChecks, totalObs)
	}

	// 4. Sort Checks: Problematic checks FIRST, then Category -> CheckName -> Target
	slices.SortFunc(checkList, func(a, b *CheckSummary) int {
		aHasIssue := (a.Failed > 0 || a.Warned > 0 || a.Unknown > 0)
		bHasIssue := (b.Failed > 0 || b.Warned > 0 || b.Unknown > 0)

		if aHasIssue && !bHasIssue {
			return -1
		}
		if !aHasIssue && bHasIssue {
			return 1
		}

		if a.Category != b.Category {
			return strings.Compare(string(a.Category), string(b.Category))
		}
		if a.CheckName != b.CheckName {
			return strings.Compare(a.CheckName, b.CheckName)
		}
		return strings.Compare(a.Target, b.Target)
	})

	// 5. Render Per-Check Details
	for _, cs := range checkList {
		// Category / CheckName / Target
		if cs.Category != "" {
			fmt.Fprintf(&b, "%s / %s / %s\n", cs.Category, cs.CheckName, cs.Target)
		} else {
			fmt.Fprintf(&b, "%s / %s\n", cs.CheckName, cs.Target)
		}

		// Observations (includes unknown if > 0)
		if cs.Unknown > 0 {
			fmt.Fprintf(&b, "  %d observations   %d pass   %d warn   %d fail   %d unknown\n",
				cs.Total, cs.Passed, cs.Warned, cs.Failed, cs.Unknown)
		} else {
			fmt.Fprintf(&b, "  %d observations   %d pass   %d warn   %d fail\n",
				cs.Total, cs.Passed, cs.Warned, cs.Failed)
		}

		// Latency Distribution
		if cs.LatencySamples > 0 {
			fmt.Fprintf(&b, "  latency  p50 %s   p95 %s   max %s   (%d samples)\n",
				formatDuration(cs.LatencyP50),
				formatDuration(cs.LatencyP95),
				formatDuration(cs.MaxLatency),
				cs.LatencySamples,
			)
		} else {
			b.WriteString("  latency  no successful samples\n")
		}

		// Failure Windows
		if len(cs.Failures) > 0 {
			for _, win := range cs.Failures {
				fmt.Fprintf(&b, "  failures %s\n", formatWindow(win))
			}
		}

		b.WriteString("\n")
	}

	_, err := fmt.Fprint(w, b.String())
	return err
}

func WriteJSONSummary(w io.Writer, s *Summary) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func WriteJSONSummaryFile(path string, s *Summary) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return WriteJSONSummary(f, s)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func formatWindow(w Window) string {
	startStr := w.Start.Format("2006-01-02 15:04:05")
	var endStr string

	if w.Start.Format("2006-01-02") == w.End.Format("2006-01-02") {
		endStr = w.End.Format("15:04:05")
	} else {
		endStr = w.End.Format("2006-01-02 15:04:05")
	}

	if w.Count == 1 {
		return fmt.Sprintf("%s (single failure)", startStr)
	}
	return fmt.Sprintf("%s → %s  (%d consecutive)", startStr, endStr, w.Count)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0ms"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

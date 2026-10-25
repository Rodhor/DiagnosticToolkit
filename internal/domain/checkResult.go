package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type ResultStatus string

const (
	Pass    ResultStatus = "Pass"
	Warn    ResultStatus = "Warn"
	Fail    ResultStatus = "Fail"
	Unknown ResultStatus = "Unknown"
)

type CheckResult struct {
	CheckType            CheckType     `json:"check_type"`
	CheckName            string        `json:"check_name"`
	Category             Category      `json:"category"`
	Target               string        `json:"target"`
	Status               ResultStatus  `json:"status"`
	Latency              time.Duration `json:"-"`
	SkipLatencyThreshold bool          `json:"-"`
	Reason               string        `json:"reason"`
	Err                  string        `json:"err"`
	Timestamp            time.Time     `json:"timestamp"`
	Metrics              []Metric      `json:"metrics"`
}

func (r CheckResult) Key() string {
	return fmt.Sprintf("%s:%s:%s", r.Category, r.CheckName, r.Target)
}

func (r CheckResult) MarshalJSON() ([]byte, error) {
	type Alias CheckResult
	ts := r.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	return json.Marshal(&struct {
		Alias
		LatencyMs int64     `json:"latency_ms"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Alias:     Alias(r),
		LatencyMs: r.Latency.Milliseconds(),
		Timestamp: ts,
	})
}

type CheckResultBuilder struct {
	result CheckResult
}

func NewResultBuilder(checkType CheckType, target string) *CheckResultBuilder {
	return &CheckResultBuilder{
		result: CheckResult{
			CheckType: checkType,
			Target:    target,
			Status:    Unknown,
			Timestamp: time.Now(),
			Metrics:   []Metric{},
		},
	}
}

func (b *CheckResultBuilder) Status(s ResultStatus) *CheckResultBuilder {
	b.result.Status = s
	return b
}

func (b *CheckResultBuilder) Latency(d time.Duration) *CheckResultBuilder {
	b.result.Latency = d
	return b
}
func (b *CheckResultBuilder) SkipLatencyThreshold() *CheckResultBuilder {
	b.result.SkipLatencyThreshold = true
	return b
}
func (b *CheckResultBuilder) Reason(r string) *CheckResultBuilder {
	b.result.Reason = r
	return b
}
func (b *CheckResultBuilder) Err(err error) *CheckResultBuilder {
	if err != nil {
		b.result.Err = err.Error()
	}
	return b
}
func (b *CheckResultBuilder) Metrics(m []Metric) *CheckResultBuilder {
	if m != nil {
		b.result.Metrics = m
	}
	return b
}
func (b *CheckResultBuilder) Metric(m Metric) *CheckResultBuilder {
	b.result.Metrics = append(b.result.Metrics, m)
	return b
}

func (b *CheckResultBuilder) Timestamp(t time.Time) *CheckResultBuilder {
	b.result.Timestamp = t
	return b
}
func (b *CheckResultBuilder) Build() CheckResult {
	return b.result
}

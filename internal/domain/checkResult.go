package domain

import (
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
	CheckType CheckType
	Target    string
	Status    ResultStatus
	Latency   time.Duration
	Reason    string
	Err       string
	Timestamp time.Time
	Metrics   []Metric
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

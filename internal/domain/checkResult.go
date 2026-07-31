package domain

import "time"

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
	Err       string
	Timestamp time.Time
	Detail    map[string]any
}

func NewCheckResult(
	checkType CheckType,
	target string,
	status ResultStatus,
	latency time.Duration,
	timestamp time.Time,
	err error,
	detail map[string]any,
) CheckResult {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	if detail == nil {
		detail = make(map[string]any)
	}
	return CheckResult{
		CheckType: checkType,
		Target:    target,
		Status:    status,
		Latency:   latency,
		Err:       errStr,
		Timestamp: timestamp,
		Detail:    detail,
	}
}

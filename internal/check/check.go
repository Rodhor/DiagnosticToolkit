package check

import (
	"context"
	"time"
)

type CheckType string

const (
	TCP  CheckType = "TCP"
	Ping CheckType = "Ping"
)

type ResultStatus string

const (
	Pass    ResultStatus = "Pass"
	Warn    ResultStatus = "Warn"
	Fail    ResultStatus = "Fail"
	Unknown ResultStatus = "Unknown"
)

type Result struct {
	CheckType CheckType
	Target    string
	Status    ResultStatus
	Latency   time.Duration
	Err       string
	Timestamp time.Time
	Detail    map[string]any
}

type Check interface {
	Run(context.Context) []Result
}

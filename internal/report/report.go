package report

import (
	"github.com/rodhor/DiagnosticToolset/internal/check"
)

type BatchReporter interface {
	WriteReport(results []check.Result) error
}

type StreamReporter interface {
	WriteResult(res check.Result) error
}

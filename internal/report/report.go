package report

import (
	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type BatchReporter interface {
	WriteReport(results []domain.CheckResult) error
}

type StreamReporter interface {
	WriteResult(res domain.CheckResult) error
}

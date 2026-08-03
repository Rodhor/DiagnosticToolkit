package report

import (
	"errors"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

type MultiStreamReporter struct {
	reporters []StreamReporter
}

func NewMultiStreamReporter(reporters ...StreamReporter) *MultiStreamReporter {
	return &MultiStreamReporter{
		reporters: reporters,
	}
}

func (m *MultiStreamReporter) WriteResult(res domain.CheckResult) error {
	var errs []error
	for _, r := range m.reporters {
		if err := r.WriteResult(res); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

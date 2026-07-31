package report

import (
	"fmt"
	"io"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
)

var (
	_ StreamReporter = (*ConsoleReport)(nil)
)

type ConsoleReport struct {
	writer io.Writer
}

func NewConsoleReport(writer io.Writer) *ConsoleReport {
	return &ConsoleReport{
		writer: writer,
	}
}

func (cr *ConsoleReport) WriteResult(result domain.CheckResult) error {
	var statusSymbol string
	switch result.Status {
	case domain.Pass:
		statusSymbol = "[PASS]"
	case domain.Fail:
		statusSymbol = "[FAIL]"
	case domain.Warn:
		statusSymbol = "[WARN]"
	default:
		statusSymbol = "[UNKNOWN]"
	}

	line := fmt.Sprintf("%s [%s] %s", statusSymbol, result.CheckType, result.Target)
	if result.Status == domain.Pass || result.Status == domain.Warn {
		line += fmt.Sprintf(" - %v", result.Latency.Round(time.Microsecond))
	}
	if reason, ok := result.Detail[domain.KeyReason]; ok {
		line += fmt.Sprintf(" [%v]", reason)
	}
	if result.Err != "" {
		line += fmt.Sprintf(" - %s", result.Err)
	}
	_, err := fmt.Fprintln(cr.writer, line)
	return err
}

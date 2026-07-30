package report

import (
	"fmt"
	"io"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/check"
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

func (cr *ConsoleReport) WriteResult(result check.Result) error {
	var statusSymbol string
	switch result.Status {
	case check.Pass:
		statusSymbol = "[PASS]"
	case check.Fail:
		statusSymbol = "[FAIL]"
	case check.Warn:
		statusSymbol = "[WARN]"
	default:
		statusSymbol = "[UNKNOWN]"
	}

	line := fmt.Sprintf("%s [%s] %s", statusSymbol, result.CheckType, result.Target)
	if result.Status == check.Pass || result.Status == check.Warn {
		line += fmt.Sprintf(" - %v", result.Latency.Round(time.Microsecond))
	}
	if reason, ok := result.Detail[check.KeyReason]; ok {
		line += fmt.Sprintf(" [%v]", reason)
	}
	if result.Err != "" {
		line += fmt.Sprintf(" - %s", result.Err)
	}
	_, err := fmt.Fprintln(cr.writer, line)
	return err
}

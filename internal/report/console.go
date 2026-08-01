package report

import (
	"fmt"
	"io"
	"strings"
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

	var line strings.Builder
	fmt.Fprintf(&line, "%s [%s] %s", statusSymbol, result.CheckName, result.Target)
	if result.Status == domain.Pass || result.Status == domain.Warn {
		fmt.Fprintf(&line, " - %v", result.Latency.Round(time.Microsecond))
	}
	if result.Reason != "" {
		fmt.Fprintf(&line, " [%s]", result.Reason)
	}
	for _, m := range result.Metrics {
		if m.Primary {
			fmt.Fprintf(&line, " %s=%s", m.Label, m.Display)
		}
	}
	if result.Err != "" {
		fmt.Fprintf(&line, " - %s", result.Err)
	}
	_, err := fmt.Fprintln(cr.writer, line.String())
	return err
}

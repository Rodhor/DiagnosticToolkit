package report

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"golang.org/x/term"
)

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiDim    = "\033[2m"
	ansiYellow = "\033[93m"
)

var (
	_ StreamReporter = (*ConsoleReport)(nil)
)

type ConsoleReport struct {
	writer io.Writer
	color  bool
}

func NewConsoleReport(writer io.Writer) *ConsoleReport {
	return &ConsoleReport{
		writer: writer,
		color:  shouldColor(writer),
	}
}

func shouldColor(w io.Writer) bool {
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return false
	}
	f, ok := w.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return false
	}
	return enableVirtualTerminal(f)
}

func (cr *ConsoleReport) colorize(s, code string) string {
	if !cr.color {
		return s
	}
	return code + s + ansiReset
}

func (cr *ConsoleReport) WriteResult(result domain.CheckResult) error {
	var statusSymbol string
	switch result.Status {
	case domain.Pass:
		statusSymbol = "[PASS]   "
	case domain.Fail:
		statusSymbol = "[FAIL]   "
	case domain.Warn:
		statusSymbol = "[WARN]   "
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
	out := line.String()
	switch result.Status {
	case domain.Fail:
		out = cr.colorize(out, ansiRed)
	case domain.Warn:
		out = cr.colorize(out, ansiYellow)
	case domain.Unknown:
		out = cr.colorize(out, ansiDim)
	}

	_, err := fmt.Fprintln(cr.writer, out)
	return err
}

const sectionWidth = 72

func writeSection(b *strings.Builder, title string) {
	fmt.Fprintf(b, "\n%s\n", strings.Repeat("━", sectionWidth))
	fmt.Fprintf(b, "  %s\n", strings.ToUpper(title))
	fmt.Fprintf(b, "%s\n\n", strings.Repeat("━", sectionWidth))
}

func writeSubSection(b *strings.Builder, title string) {
	fmt.Fprintf(b, "\n%s\n%s\n", title, strings.Repeat("─", len(title)))
}

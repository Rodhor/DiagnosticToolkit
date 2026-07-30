package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/report"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	targets := []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
		"google.com:80",
		"127.0.0.1:9999",
	}
	timeout := 3 * time.Second
	warnThreshold := 300 * time.Millisecond
	dialer := &net.Dialer{}

	checks := []check.Check{
		check.NewTCPCheck(targets, timeout, warnThreshold, dialer),
	}

	reporter := report.NewConsoleReport(os.Stdout)
	var allResults []check.Result
	hasFailure := false
	hasWarning := false

	for _, c := range checks {
		results := c.Run(ctx)
		for _, res := range results {
			allResults = append(allResults, res)
			_ = reporter.WriteResult(res)

			if res.Status == check.Fail {
				hasFailure = true
			}
			if res.Status == check.Warn {
				hasWarning = true
			}
		}
	}

	if hasFailure {
		return 1
	}
	if hasWarning {
		return 2
	}
	return 0
}

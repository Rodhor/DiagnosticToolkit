package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/check"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/report"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	tcpTargets := []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
		"google.com:80",
		"127.0.0.1:9999",
	}

	resolvers := []check.ResolverSpec{
		{Name: "System", Address: ""},
		{Name: "Cloudflare", Address: "1.1.1.1:53"},
		{Name: "Google", Address: "8.8.8.8:53"},
		{Name: "UnreachableLocal", Address: "127.0.0.1:9999"},
	}

	queries := []check.QuerySpec{
		{
			Name:              "one.one.one.one",
			ExpectedOutcome:   check.ExpectedSuccess,
			ExpectedAddresses: []string{"1.1.1.1"},
		},
		{
			Name:              "dns.google",
			ExpectedOutcome:   check.ExpectedSuccess,
			ExpectedAddresses: []string{"8.8.8.8"},
		},
		{
			Name:            "google.com",
			ExpectedOutcome: check.ExpectedSuccess,
		},
		{
			Name:            "this-domain-definitely-does-not-exist-12345.com",
			ExpectedOutcome: check.ExpectedNXDomain,
		},
	}

	thresholds := domain.Thresholds{
		domain.TCP: 300 * time.Millisecond,
		domain.DNS: 200 * time.Millisecond,
	}
	timeout := 3 * time.Second
	dialer := &net.Dialer{}

	checks := []check.Check{
		check.NewTCPCheck(tcpTargets, timeout, dialer),
		check.NewDNSCheck(resolvers, queries, timeout),
	}

	reporter := report.NewConsoleReport(os.Stdout)
	hasFailure := false
	hasWarning := false

	for _, c := range checks {
		// Apply thresholds to the results of each check
		results := c.Run(ctx)

		results = domain.ApplyThresholds(results, thresholds)

		for _, res := range results {
			_ = reporter.WriteResult(res)

			if res.Status == domain.Fail {
				hasFailure = true
			}
			if res.Status == domain.Warn {
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

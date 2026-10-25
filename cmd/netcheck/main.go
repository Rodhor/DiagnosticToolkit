package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/report"
)

func main() {
	os.Exit(run())
}

func run() int {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	monitor := flag.Bool("monitor", false, "run continuously until interrupted")
	flag.BoolVar(monitor, "m", false, "shorthand for -monitor")
	output := flag.String("output", "", "write results as JSON Lines to this file")
	summaryPath := flag.String("summary", "", "write the run summary as JSON to this file")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := app.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		return 1
	}

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "monitor", "m":
			cfg.Monitor = *monitor
		case "output":
			cfg.OutputPath = *output
		case "summary":
			cfg.SummaryPath = *summaryPath
		}
	})

	console := report.NewConsoleReport(os.Stdout)
	reporters := []report.StreamReporter{console}
	if cfg.OutputPath != "" {
		jsonl, err := report.NewJSONLFileReporter(cfg.OutputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "output file: %v", err)
			return 1
		}
		defer jsonl.Close()
		reporters = append(reporters, jsonl)
	}

	reporter := report.NewMultiStreamReporter(reporters...)

	summary := report.NewSummary(time.Now())
	defer func() {
		summary.Finalize(time.Now())
		if err := report.WriteSummary(os.Stdout, summary); err != nil {
			fmt.Fprintf(os.Stderr, "writing summary: %v\n", err)
		}
		if cfg.SummaryPath != "" {
			if err := report.WriteJSONSummaryFile(cfg.SummaryPath, summary); err != nil {
				fmt.Fprintf(os.Stderr, "writing summary file: %v\n", err)
			}
		}
	}()

	a, err := app.New(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup: %v\n", err)
		return 1
	}

	if snap, snapErr := a.Environment(); snapErr != nil {
		fmt.Fprintf(os.Stderr, "environment discovery failed: %v\n", snapErr)
	} else {
		_ = console.WriteEnvironment(snap)
	}

	out := make(chan domain.CheckResult, 64)
	done := make(chan struct{})
	go func() {
		defer close(done)
		consume(out, reporter, summary)
	}()

	var results []domain.CheckResult
	if cfg.Monitor {
		a.Monitor(ctx, out)
	} else {
		results = a.RunOnce(ctx, out)
	}
	<-done

	if cfg.Monitor {
		return exitCodeFromSummary(summary)
	}
	return exitCode(results)
}

func exitCode(results []domain.CheckResult) int {
	warned := false
	for _, r := range results {
		switch r.Status {
		case domain.Fail:
			return 1
		case domain.Warn:
			warned = true
		}
	}
	if warned {
		return 2
	}
	return 0
}

func exitCodeFromSummary(s *report.Summary) int {
	exitCode := 0
	for _, cs := range s.Checks {
		if cs.Warned > 0 {
			exitCode = 2
		}
		if cs.Failed > 0 {
			exitCode = 1
		}
	}
	return exitCode
}

func defaultConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "netcheck.yaml"
	}
	return filepath.Join(filepath.Dir(exe), "netcheck.yaml")
}

func consume(
	in <-chan domain.CheckResult,
	rep report.StreamReporter,
	sum *report.Summary,
) {
	for res := range in {
		sum.Add(res)
		if err := rep.WriteResult(res); err != nil {
			fmt.Fprintf(os.Stderr, "writing result: %v\n", err)
		}
	}
}

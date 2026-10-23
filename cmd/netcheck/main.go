package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/report"
)

func main() {
	os.Exit(run())
}

func run() int {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := app.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		return 1
	}
	a, err := app.New(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup failed: %v\n\n", err)
		return 1
	}

	reporter := report.NewConsoleReport(os.Stdout)

	snap, snapErr := a.Environment()
	if snapErr != nil {
		fmt.Fprintf(os.Stderr, "environment discovery failed: %v\n", snapErr)
	} else {
		_ = reporter.WriteEnvironment(snap)
	}

	results := a.RunOnce(ctx, nil)
	for _, r := range results {
		_ = reporter.WriteResult(r)
	}

	return exitCode(results)
}

func exitCode(results []domain.CheckResult) int {
	for _, r := range results {
		switch r.Status {
		case domain.Fail:
			return 1
		case domain.Warn:
			return 2
		}
	}
	return 0
}

func defaultConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "netcheck.yaml"
	}
	return filepath.Join(filepath.Dir(exe), "netcheck.yaml")
}

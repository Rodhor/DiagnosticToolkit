package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/helper"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
)

func main() {
	os.Exit(run())
}

func run() int {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	addrFlag := flag.String("addr", helper.DefaultAddr(), "listen address (HOST:PORT)")
	portFlag := flag.Int("port", 0, "port to listen on (127.0.0.1:<port>)")
	flag.Parse()

	addr := *addrFlag
	if *portFlag != 0 {
		if *portFlag < 1 || *portFlag > 65535 {
			fmt.Fprintf(os.Stderr, "error: port %d is out of range (must be between 1 and 65535)\n", *portFlag)
			return 1
		}
		addr = fmt.Sprintf("127.0.0.1:%d", *portFlag)
	} else {
		var err error
		addr, err = helper.ParseAddr(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	}

	if err := helper.IsAddrAvailable(addr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	if err := runServer(ctx, a, addr); err != nil {
		slog.Error("server error", "err", err)
		return 1
	}
	return 0
}

func runServer(ctx context.Context, a *app.App, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           server.New(a).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("gracefull shutdown failed", "err", err)
		}
	}()

	slog.Info("starting server", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func defaultConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "netcheck.yaml"
	}
	return filepath.Join(filepath.Dir(exe), "netcheck.yaml")
}

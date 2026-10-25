package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
	"github.com/rodhor/DiagnosticToolset/web/templates/components"
)

func Start(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.State().Start(r.Context())
		templ.Handler(components.Status(s.State().IsRunning())).ServeHTTP(w, r)
	}
}

func Stop(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.State().Stop()
		w.Header().Set("HX-Redirect", "/summary")
		w.WriteHeader(http.StatusNoContent)
	}
}

func RunOnce(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.State().RunOnce(r.Context())
		templ.Handler(components.Results(s.State().Recent())).ServeHTTP(w, r)
	}
}

func ReloadConfig(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func SaveConfig(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		v := app.SettingsView{
			Timeout:        r.FormValue("timeout"),
			Concurrency:    atoi(r.FormValue("concurrency")),
			Interval:       r.FormValue("interval"),
			ThresholdTCP:   r.FormValue("threshold_tcp"),
			ThresholdDNS:   r.FormValue("threshold_dns"),
			ThresholdPing:  r.FormValue("threshold_ping"),
			PingCount:      atoi(r.FormValue("ping_count")),
			PingInterval:   r.FormValue("ping_interval"),
			PingTimeout:    r.FormValue("ping_timeout"),
			PingWarnLoss:   parseFloatOr(r.FormValue("ping_warn_loss"), 0),
			PingFailLoss:   parseFloatOr(r.FormValue("ping_fail_loss"), 0),
			TCPTargets:     r.Form["tcp_targets"],
			PingTargets:    r.Form["ping_targets"],
			Resolvers:      r.Form["resolvers"],
			OutputResults:  r.FormValue("output_results"),
			OutputSummary:  r.FormValue("output_summary"),
		}
		// Parse indexed query rows.
		for i := 0; i < 32; i++ {
			name := strings.TrimSpace(r.FormValue(fmt.Sprintf("q_name_%d", i)))
			if name == "" {
				continue
			}
			v.Queries = append(v.Queries, app.QueryView{
				Name:     name,
				Expected: r.FormValue(fmt.Sprintf("q_expected_%d", i)),
				Expect:   r.FormValue(fmt.Sprintf("q_expect_%d", i)),
			})
		}

		cfg, err := app.ViewToConfig(v, app.Default())
		if err != nil {
			http.Error(w, "invalid config: "+err.Error(), http.StatusBadRequest)
			return
		}

		cfgPath := s.State().ConfigPath()
		// Backup current file.
		if data, err := os.ReadFile(cfgPath); err == nil {
			_ = os.WriteFile(cfgPath+".bak", data, 0o644)
		}

		if err := app.SaveConfig(cfgPath, cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Reload app.
		if newCfg, err := app.Load(cfgPath); err == nil {
			if a, err := app.New(context.Background(), newCfg); err == nil {
				s.State().SetApp(a)
			}
		}

		http.Redirect(w, r, "/settings?saved=1", http.StatusSeeOther)
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func parseFloatOr(s string, def float64) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return def
	}
	return f
}

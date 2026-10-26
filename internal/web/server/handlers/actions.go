package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
	"github.com/rodhor/DiagnosticToolset/web/templates/components"
	"github.com/rodhor/DiagnosticToolset/web/templates/pages"
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
		results := s.State().RunOnce(r.Context())
		templ.Handler(components.Results(results)).ServeHTTP(w, r)
	}
}

func SaveConfig(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			renderSettingsError(w, r, app.ConfigToView(app.Default()), "The form could not be read. Please review the settings and try again.", http.StatusBadRequest)
			return
		}

		v := app.SettingsView{
			Timeout:       r.FormValue("timeout"),
			Concurrency:   r.FormValue("concurrency"),
			Interval:      r.FormValue("interval"),
			ThresholdTCP:  r.FormValue("threshold_tcp"),
			ThresholdDNS:  r.FormValue("threshold_dns"),
			ThresholdPing: r.FormValue("threshold_ping"),
			PingCount:     r.FormValue("ping_count"),
			PingInterval:  r.FormValue("ping_interval"),
			PingTimeout:   r.FormValue("ping_timeout"),
			PingWarnLoss:  r.FormValue("ping_warn_loss"),
			PingFailLoss:  r.FormValue("ping_fail_loss"),
			TCPTargets:    r.Form["tcp_targets"],
			PingTargets:   r.Form["ping_targets"],
			Resolvers:     r.Form["resolvers"],
			OutputResults: r.FormValue("output_results"),
			OutputSummary: r.FormValue("output_summary"),
		}
		if err := readQueryRows(r, &v); err != nil {
			renderSettingsError(w, r, v, err.Error(), http.StatusBadRequest)
			return
		}

		cfg, err := app.ViewToConfig(v, app.Default())
		if err != nil {
			renderSettingsError(w, r, v, err.Error(), http.StatusBadRequest)
			return
		}
		validationPath, cleanup, err := validateConfigTemp(cfg)
		if err != nil {
			renderSettingsError(w, r, v, "Could not validate settings: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer cleanup()
		if _, err := app.Load(validationPath); err != nil {
			renderSettingsError(w, r, v, "Invalid configuration: "+err.Error(), http.StatusBadRequest)
			return
		}

		cfgPath := s.State().ConfigPath()
		if data, err := os.ReadFile(cfgPath); err == nil {
			if err := os.WriteFile(cfgPath+".bak", data, 0o644); err != nil {
				renderSettingsError(w, r, v, "Could not back up the current config: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			renderSettingsError(w, r, v, "Could not read the current config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := app.SaveConfig(cfgPath, cfg); err != nil {
			renderSettingsError(w, r, v, "Could not save settings: "+err.Error(), http.StatusInternalServerError)
			return
		}

		newCfg, err := app.Load(cfgPath)
		if err != nil {
			renderSettingsError(w, r, v, "Settings were saved but could not be reloaded: "+err.Error(), http.StatusInternalServerError)
			return
		}
		a, err := app.New(context.Background(), newCfg)
		if err != nil {
			renderSettingsError(w, r, v, "Settings were saved but could not be reloaded: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.State().SetApp(a)
		http.Redirect(w, r, "/settings?saved=1", http.StatusSeeOther)
	}
}

func readQueryRows(r *http.Request, view *app.SettingsView) error {
	var indexes []int
	for key := range r.Form {
		if !strings.HasPrefix(key, "q_name_") {
			continue
		}
		index, err := strconv.Atoi(strings.TrimPrefix(key, "q_name_"))
		if err != nil || index < 0 {
			return fmt.Errorf("invalid DNS query row")
		}
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		name := strings.TrimSpace(r.FormValue(fmt.Sprintf("q_name_%d", index)))
		if name == "" {
			continue
		}
		view.Queries = append(view.Queries, app.QueryView{
			Name:     name,
			Expected: r.FormValue(fmt.Sprintf("q_expected_%d", index)),
			Expect:   r.FormValue(fmt.Sprintf("q_expect_%d", index)),
		})
	}
	return nil
}

func validateConfigTemp(cfg app.Config) (string, func(), error) {
	tmp, err := os.CreateTemp("", "diagnostic-settings-*.yaml")
	if err != nil {
		return "", nil, err
	}
	path := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(path)
		return "", nil, err
	}
	if err := os.Remove(path); err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_ = os.Remove(path)
		_ = os.Remove(path + ".tmp")
	}
	if err := app.SaveConfig(path, cfg); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}

func renderSettingsError(w http.ResponseWriter, r *http.Request, view app.SettingsView, message string, status int) {
	w.WriteHeader(status)
	templ.Handler(pages.Settings(view, false, "", message)).ServeHTTP(w, r)
}

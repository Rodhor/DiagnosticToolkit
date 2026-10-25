package handlers

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/rodhor/DiagnosticToolset/internal/app"
	"github.com/rodhor/DiagnosticToolset/internal/report"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
	"github.com/rodhor/DiagnosticToolset/web/templates/pages"
)

func Dashboard(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runs := s.State().Runs()
		// newest first
		rev := make([]*server.Run, len(runs))
		for i, r := range runs {
			rev[len(runs)-1-i] = r
		}
		templ.Handler(pages.Dashboard(s.State().IsRunning(), rev)).ServeHTTP(w, r)
	}
}

func Logs(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := 0
		if v := r.URL.Query().Get("run"); v != "" {
			fmt.Sscanf(v, "%d", &runID)
		}
		templ.Handler(pages.Logs(runID)).ServeHTTP(w, r)
	}
}

func Settings(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, _ := app.Load(s.State().ConfigPath())
		view := app.ConfigToView(cfg)
		saved := r.URL.Query().Get("saved") == "1"
		templ.Handler(pages.Settings(view, saved)).ServeHTTP(w, r)
	}
}

func Environment(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap, err := s.State().App().Environment()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Handler(pages.Environment(snap)).ServeHTTP(w, r)
	}
}

func Summary(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := 0
		if v := r.URL.Query().Get("run"); v != "" {
			fmt.Sscanf(v, "%d", &runID)
		}
		var sum *report.Summary
		if runID > 0 {
			if r := s.State().RunByID(runID); r != nil {
				sum = r.Summary
			}
		} else {
			sum = s.State().Summary()
		}
		if sum == nil {
			http.Error(w, "no summary available; run monitoring first", http.StatusNotFound)
			return
		}
		templ.Handler(pages.Summary(sum)).ServeHTTP(w, r)
	}
}

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
		snapshot, err := s.State().App().Environment()
		snapshotErr := ""
		if err != nil {
			snapshotErr = err.Error()
		}
		templ.Handler(pages.Dashboard(s.State().IsRunning(), newestFirst(s.State().Runs()), snapshot, snapshotErr)).ServeHTTP(w, r)
	}
}

func History(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(pages.History(newestFirst(s.State().Runs()))).ServeHTTP(w, r)
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
		cfg, err := app.Load(s.State().ConfigPath())
		loadError := ""
		if err != nil {
			loadError = err.Error()
			cfg = app.Default()
		}
		view := app.ConfigToView(cfg)
		saved := r.URL.Query().Get("saved") == "1"
		templ.Handler(pages.Settings(view, saved, loadError, "")).ServeHTTP(w, r)
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
		var run *server.Run
		if runID > 0 {
			run = s.State().RunByID(runID)
		} else {
			runs := s.State().Runs()
			if len(runs) > 0 {
				run = runs[len(runs)-1]
			}
		}
		if run != nil && !run.Running {
			sum = run.Summary
		}
		templ.Handler(pages.Summary(sum, runID, run != nil && run.Running)).ServeHTTP(w, r)
	}
}

func newestFirst(runs []*server.Run) []*server.Run {
	for i, j := 0, len(runs)-1; i < j; i, j = i+1, j-1 {
		runs[i], runs[j] = runs[j], runs[i]
	}
	return runs
}

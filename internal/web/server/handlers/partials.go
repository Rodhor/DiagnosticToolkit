package handlers

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/rodhor/DiagnosticToolset/internal/domain"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
	"github.com/rodhor/DiagnosticToolset/web/templates/components"
)

func PartialStatus(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(components.Status(s.State().IsRunning())).ServeHTTP(w, r)
	}
}

func PartialResults(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := 0
		if v := r.URL.Query().Get("run"); v != "" {
			fmt.Sscanf(v, "%d", &runID)
		}
		var items []domain.CheckResult
		if runID > 0 {
			if run := s.State().RunByID(runID); run != nil {
				items = run.Buffer.All()
			}
		} else {
			items = s.State().Recent()
		}
		if len(items) == 0 && runID == 0 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<p class="text-muted-foreground">No active run.</p>`))
			return
		}
		templ.Handler(components.Results(items)).ServeHTTP(w, r)
	}
}

func PartialEnvironment(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap, err := s.State().App().Environment()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Handler(components.Environment(snap)).ServeHTTP(w, r)
	}
}

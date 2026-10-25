package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
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
		b, _ := json.Marshal(s.State().Recent())
		templ.Handler(components.Results(string(b))).ServeHTTP(w, r)
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

package handlers

import (
	"net/http"

	"github.com/a-h/templ"
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

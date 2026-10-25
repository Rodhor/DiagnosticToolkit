package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
	"github.com/rodhor/DiagnosticToolset/web/templates/pages"
)

func Dashboard(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(pages.Dashboard(s.State().IsRunning())).ServeHTTP(w, r)
	}
}

func Logs(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(pages.Logs()).ServeHTTP(w, r)
	}
}

func Settings(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(pages.Settings()).ServeHTTP(w, r)
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
		sum := s.State().Summary()
		if sum == nil {
			http.Error(w, "no summary available; run monitoring first", http.StatusNotFound)
			return
		}
		templ.Handler(pages.Summary(sum)).ServeHTTP(w, r)
	}
}

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rodhor/DiagnosticToolset/internal/web/server"
)

func Start(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.State().Start(r.Context()) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusConflict)
	}
}

func Stop(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.State().Stop()
		w.WriteHeader(http.StatusNoContent)
	}
}

func RunOnce(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := s.State().RunOnce(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func ReloadConfig(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

package handlers

import (
	"net/http"

	"github.com/rodhor/DiagnosticToolset/internal/web/server"
)

func SSE(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		client := s.State().Hub().Subscribe()
		defer s.State().Hub().Unsubscribe(client)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case msg, ok := <-client.Ch:
				if !ok {
					return
				}
				w.Write([]byte("data: "))
				w.Write(msg)
				w.Write([]byte("\n\n"))
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}

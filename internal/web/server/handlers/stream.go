package handlers

import (
	"net/http"
	"strings"

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

		// Send current results immediately on connect.
		if b, err := s.State().RenderResults(); err == nil {
			writeSSE(w, flusher, b)
		}

		for {
			select {
			case msg, ok := <-client.Ch:
				if !ok {
					return
				}
				writeSSE(w, flusher, msg)
			case <-r.Context().Done():
				return
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, data []byte) {
	for _, line := range strings.Split(string(data), "\n") {
		w.Write([]byte("data: " + line + "\n"))
	}
	w.Write([]byte("\n"))
	flusher.Flush()
}

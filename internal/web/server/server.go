package server

import (
	"net/http"

	"github.com/rodhor/DiagnosticToolset/internal/app"
)

type Server struct {
	app *app.App
}

func New(app *app.App) *Server {
	return &Server{app: app}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	setupAssetRoutes(mux)

	return mux
}

func setupAssetRoutes(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=300")
			fs.ServeHTTP(w, r)
		})))
}

package handlers

import (
	"net/http"

	"github.com/templui/templui/utils"
	"github.com/rodhor/DiagnosticToolset/internal/web/server"
)

func Routes(s *server.Server) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /", Dashboard(s))
	mux.Handle("GET /logs", Logs(s))
	mux.Handle("GET /settings", Settings(s))
	mux.Handle("GET /environment", Environment(s))
	mux.Handle("GET /summary", Summary(s))

	mux.Handle("POST /actions/start", Start(s))
	mux.Handle("POST /actions/stop", Stop(s))
	mux.Handle("POST /actions/run-once", RunOnce(s))
	mux.Handle("POST /actions/reload-config", ReloadConfig(s))

	mux.Handle("GET /stream", SSE(s))

	mux.Handle("GET /partials/status", PartialStatus(s))
	mux.Handle("GET /partials/results", PartialResults(s))
	mux.Handle("GET /partials/environment", PartialEnvironment(s))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	utils.SetupScriptRoutes(mux, true)

	return mux
}

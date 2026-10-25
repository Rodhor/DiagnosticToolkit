package server

import (
	"github.com/rodhor/DiagnosticToolset/internal/app"
)

type Server struct {
	app   *app.App
	state *State
}

func New(a *app.App) *Server {
	return &Server{app: a, state: newState(a)}
}

func (s *Server) State() *State { return s.state }

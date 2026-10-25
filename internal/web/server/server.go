package server

import (
	"github.com/rodhor/DiagnosticToolset/internal/app"
)

type Server struct {
	app        *app.App
	configPath string
	state      *State
}

func New(a *app.App, configPath string) *Server {
	return &Server{app: a, configPath: configPath, state: newState(a, configPath)}
}

func (s *Server) State() *State { return s.state }

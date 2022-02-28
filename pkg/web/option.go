package web

import (
	"github.com/hashicorp/go-hclog"
)

func WithStorage(ws Storage) Option {
	return func(s *Server) { s.s = ws }
}

func WithLogger(l hclog.Logger) Option {
	return func(s *Server) { s.l = l.Named("web") }
}

func WithBasePath(path string) Option {
	return func(s *Server) { s.basePath = path }
}

func WithTemplateDebug(d bool) Option {
	return func(s *Server) { s.tmpls.Debug = d }
}

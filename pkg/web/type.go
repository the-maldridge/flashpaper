package web

import (
	"context"
	"net/http"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-hclog"
)

// Storage provides an interface to the storage system.  This can be
// provided by anything assuming it provides a mechanism for
// expiration.
type Storage interface {
	Ping(context.Context) error

	PutEx(context.Context, string, interface{}, time.Duration) error
	Get(context.Context, string) (string, error)
	Del(context.Context, string) error
}

// Server contains the various components of the flashpaper server.
type Server struct {
	l hclog.Logger
	r chi.Router
	s Storage
	n *http.Server

	basePath string

	tmpls *pongo2.TemplateSet
}

// Option is a way of configuring the server using variadic options.
// This is the preferred way to make config rather than shoving in
// lots of different types and having a bunch of different setters.
type Option func(*Server)

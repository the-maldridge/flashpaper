package web

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
)

// Server contains the various components of the flashpaper server.
type Server struct {
	l hclog.Logger
	r chi.Router
	s Storage

	tmpls *pongo2.TemplateSet
}

// New returns an initialized server instance.
func New(l hclog.Logger) (*Server, error) {
	sbl, err := pongo2.NewSandboxedFilesystemLoader("theme/p2")
	if err != nil {
		return nil, err
	}

	x := Server{
		l:     l.Named("web"),
		r:     chi.NewRouter(),
		tmpls: pongo2.NewSet("html", sbl),
	}
	x.tmpls.Debug = true

	x.r.Use(middleware.Logger)
	x.r.Use(middleware.Heartbeat("/ping"))

	x.fileServer(x.r, "/static", http.Dir("theme/static"))

	x.r.Get("/", x.index)
	x.r.Post("/paste/submit", x.acceptPaste)
	return &x, nil
}

// SetStorage allows the storage engine to be setup.
func (s *Server) SetStorage(st Storage) {
	s.s = st
}

// Serve blocks and serves.
func (s *Server) Serve(bind string) error {
	s.l.Info("Webserver is starting")
	return http.ListenAndServe(bind, s.r)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	t, err := s.tmpls.FromCache("index.p2")
	if err != nil {
		s.templateErrorHandler(w, err)
		return
	}
	if err := t.ExecuteWriter(nil, w); err != nil {
		s.templateErrorHandler(w, err)
	}
}

func (s *Server) acceptPaste(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	s.l.Debug("Submitted Form Data", "validity", r.Form.Get("validity"), "paste", r.Form.Get("paste"))

	validInterval, err := time.ParseDuration(r.Form.Get("validity"))
	if err != nil {
		validInterval = time.Minute * 15
	}

	idUUID, err := uuid.NewUUID()
	if err != nil {
		s.l.Warn("Error creating uuid", "error", err)
	}

	if err := s.s.PutEx(idUUID.String(), r.Form.Get("paste"), validInterval); err != nil {
		s.l.Warn("Error with storage", "error", err)
	}

	ctx := pongo2.Context{
		"validity":   r.Form.Get("validity"),
		"url":        path.Join("http://"+r.Host, "paste/get", idUUID.String()),
		"expiration": time.Now().Add(validInterval).Format(time.RFC850),
	}

	t, err := s.tmpls.FromCache("success.p2")
	if err != nil {
		s.templateErrorHandler(w, err)
		return
	}
	if err := t.ExecuteWriter(ctx, w); err != nil {
		s.templateErrorHandler(w, err)
	}
}

func (s *Server) templateErrorHandler(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Error while rendering template: %s\n", err)
}

func (s *Server) fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

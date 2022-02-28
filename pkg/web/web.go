package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
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
	n *http.Server

	basePath string

	tmpls *pongo2.TemplateSet
}

// New returns an initialized server instance.
func New(l hclog.Logger) (*Server, error) {
	sbl, err := pongo2.NewSandboxedFilesystemLoader("theme/p2")
	if err != nil {
		return nil, err
	}

	x := Server{
		l:        l.Named("web"),
		r:        chi.NewRouter(),
		n:        &http.Server{},
		basePath: os.Getenv("FLASHPAPER_BASEPATH"),
		tmpls:    pongo2.NewSet("html", sbl),
	}
	x.tmpls.Debug = true
	x.n.Handler = x.r

	x.r.Use(middleware.Logger)
	x.r.Use(middleware.Heartbeat("/ping"))
	x.r.Use(x.checkStorage)

	x.fileServer(x.r, path.Join(x.basePath, "/static"), http.Dir("theme/static"))

	x.r.Get(path.Join("/", x.basePath), x.index)
	x.r.Post(path.Join("/", x.basePath, "/paste/submit"), x.acceptPaste)
	x.r.Get(path.Join("/", x.basePath, "/paste/{pasteID}/{key}"), x.getPaste)
	return &x, nil
}

// SetStorage allows the storage engine to be setup.
func (s *Server) SetStorage(st Storage) {
	s.s = st
}

// Serve blocks and serves.
func (s *Server) Serve(bind string) error {
	s.l.Info("Webserver is starting")
	s.n.Addr = bind
	return s.n.ListenAndServe()
}

// Shutdown terminates the HTTP server
func (s *Server) Shutdown() {
	s.n.Close()
}

func (s *Server) checkStorage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.s.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Fatal Error: Storage service is unavailable: %v", err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	t, err := s.tmpls.FromCache("index.p2")
	if err != nil {
		s.templateErrorHandler(w, err)
		return
	}
	if err := t.ExecuteWriter(pongo2.Context{"base_path": s.basePath}, w); err != nil {
		s.templateErrorHandler(w, err)
	}
}

func (s *Server) acceptPaste(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	validInterval, err := time.ParseDuration(r.Form.Get("validity"))
	if err != nil {
		validInterval = time.Minute * 15
	}
	idUUID, err := uuid.NewUUID()
	if err != nil {
		s.l.Warn("Error creating uuid", "error", err)
	}
	id := strconv.Itoa(int(idUUID.ID()))

	data, key := encrypt(r.Form.Get("paste"))

	if err := s.s.PutEx(context.Background(), id, data, validInterval); err != nil {
		s.l.Warn("Error with storage", "error", err)
	}

	ctx := pongo2.Context{
		"base_path":  s.basePath,
		"validity":   r.Form.Get("validity"),
		"url":        path.Join("http://"+r.Host, s.basePath, "paste", id, key),
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

func (s *Server) getPaste(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "pasteID")

	paste, ferr := s.s.Get(context.Background(), id)
	if ferr != nil {
		s.l.Warn("Error retrieving paste", "error", ferr)
	}
	err := s.s.Del(context.Background(), id)
	if err != nil {
		s.l.Warn("Error deleting paste", "error", err)
	}

	// Only try to decrypt the paste if we actually got something
	// above.  Otherwise leave paste nil so that we get the error
	// page below.
	if ferr == nil {
		paste = decrypt(paste, chi.URLParam(r, "key"))
	}

	ctx := pongo2.Context{
		"base_path": s.basePath,
		"paste":     paste,
	}

	t, err := s.tmpls.FromCache("paste.p2")
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

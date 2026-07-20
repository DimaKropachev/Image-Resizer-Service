package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	s *http.Server
}

func NewServer(h Handler) *Server {
	router := chi.NewRouter()

	router.Post("api/v1/upload", h.Download)
	router.Get("api/v1/status", h.CheckStatus)
	router.Get("api/v1/download", h.Download)

	s := &http.Server{
		Handler: router,
	}

	return &Server{
		s: s,
	}
}

func (s *Server) Run() error {
	return s.s.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}

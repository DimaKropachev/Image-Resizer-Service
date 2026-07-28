package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dimakropachev/image_resizer_service/internal/config"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	s *http.Server
}

func NewServer(cfg config.HTTP, h Handler) *Server {
	router := chi.NewRouter()

	router.Post("/api/v1/upload", h.UploadImage)
	router.Get("/api/v1/status", h.CheckStatus)
	router.Get("/api/v1/download", h.Download)

	addr := fmt.Sprintf(":%d", cfg.Port)

	s := &http.Server{
		Addr:    addr,
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

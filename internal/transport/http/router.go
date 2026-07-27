package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Config struct {
	Port    int           `yaml:"port" env:"HTTP_PORT" env-default:"8088"`
	Timeout time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"5s"`
}

type Server struct {
	s *http.Server
}

func NewServer(cfg Config, h Handler) *Server {
	router := chi.NewRouter()

	router.Post("/api/v1/upload", h.UploadImage)
	router.Get("/api/v1/status", h.CheckStatus)
	router.Get("/api/v1/download", h.Download)

	addr := fmt.Sprintf(":%d", cfg.Port)

	s := &http.Server{
		Addr: addr,
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

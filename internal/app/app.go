package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dimakropachev/image_resizer_service/internal/config"
	"github.com/dimakropachev/image_resizer_service/internal/queue"
	"github.com/dimakropachev/image_resizer_service/internal/repository"
	"github.com/dimakropachev/image_resizer_service/internal/service"
	h "github.com/dimakropachev/image_resizer_service/internal/transport/http"
	"github.com/dimakropachev/image_resizer_service/internal/worker"
	wm "github.com/dimakropachev/image_resizer_service/internal/worker_manager"
)

type App struct {
	cfg *config.Config
}

func New(cfg *config.Config) (*App, error) {
	if err := createStorage(cfg.Storage); err != nil {
		return nil, err
	}

	return &App{
		cfg: cfg,
	}, nil
}

func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	repo := repository.New()
	q := queue.New(ctx)
	service := service.New(repo, q)
	wm := wm.New()
	handler := h.NewHandler(service, wm, a.cfg.Storage)
	server := h.NewServer(a.cfg.HTTP, handler)

	serverErrCh := make(chan error, 1)

	wg := sync.WaitGroup{}

	slog.Info("starting http server...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("http server started",
			slog.Int("port", a.cfg.HTTP.Port),
		)
		if err := server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
	}()

	workerErrCh := make(chan error)
	slog.Info("starting workers...")
	for i := 1; i <= a.cfg.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			slog.Info("worker start",
				slog.Int("id", id),
			)
			w := worker.New(id)
			wm.Register(id, w)
			w.Start(ctx, q, workerErrCh)
			wm.Unregister(id)
		}(i)
	}

	wg.Go(func() {
		handlerWorkerError(ctx, workerErrCh)
	})

	// ---------------------Graceful Shutdown---------------------

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sign := <-shutdownCh:
		slog.Info("Shutdown signal recieved, starting graceful shutdown...",
			slog.String("signal", sign.String()),
		)
	case serverErr := <-serverErrCh:
		slog.Info("Critical server error, initialized graceful shutdown...",
			slog.String("error", serverErr.Error()),
		)
	}
	// calling function that completes workers and closes queue
	cancel()

	shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("stopping http server...")
	if err := server.Stop(shutdownContext); err != nil {
		slog.Error("error shutdown http server",
			slog.String("error", err.Error()),
		)
	} else {
		slog.Info("http server stopped")
	}

	wg.Wait()
}

func handlerWorkerError(ctx context.Context, wErr <-chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-wErr:
			if !ok {
				return
			}
			slog.Warn("worker error",
				slog.String("error", err.Error()),
			)
		}
	}
}

func createStorage(paths config.Storage) error {
	if err := os.MkdirAll(paths.UploadPath, 0755); err != nil {
		return fmt.Errorf("couldn't create dir for upload img: %w", err)
	}
	if err := os.MkdirAll(paths.ProcessedPath, 0755); err != nil {
		return fmt.Errorf("couldn't create dir for processed img: %w", err)
	}
	return nil
}

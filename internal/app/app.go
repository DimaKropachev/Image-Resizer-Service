package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dimakropachev/image_resizer_service/internal/config"
	"github.com/dimakropachev/image_resizer_service/internal/queue"
	"github.com/dimakropachev/image_resizer_service/internal/repository"
	"github.com/dimakropachev/image_resizer_service/internal/service"
	"github.com/dimakropachev/image_resizer_service/internal/transport/http"
	"github.com/dimakropachev/image_resizer_service/internal/worker"
)

type App struct {
	cfg *config.Config
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	repo := repository.New()
	service := service.New(repo)
	handler := http.NewHandler(service)
	server := http.NewServer(handler)

	serverErrCh := make(chan error, 1)

	wg := sync.WaitGroup{}

	slog.Info("starting http server...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("http server started", slog.Int("port", a.cfg.HTTP.Port))
		if err := server.Run(); err != nil {
			serverErrCh <- err
			return
		}
	}()

	q := queue.New(ctx)
	workerErrCh := make(chan error)
	slog.Info("starting workers...")
	for i := 1; i <= a.cfg.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			slog.Info("worker start", slog.Int("id", id))
			w := worker.New(id)
			w.Start(ctx, q, workerErrCh)
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		handlerWorkerError(ctx, workerErrCh)
	}()

	// ---------------------Graceful Shutdown---------------------

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sign := <-shutdownCh:
		slog.Info("Shutdown signal recieved, starting graceful shutdown...", slog.String("signal", sign.String()))
	case serverErr := <-serverErrCh:
		slog.Info("Critical server error, initialized graceful shutdown...", slog.String("error", serverErr.Error()))
	}
	// calling function that completes workers and closes queue
	cancel()

	shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("stopping http server...")
	if err := server.Stop(shutdownContext); err != nil {
		slog.Error("error shutdown http server", slog.String("error", err.Error()))
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
			slog.Debug("worker error", slog.String("error", err.Error()))
		}
	}
}

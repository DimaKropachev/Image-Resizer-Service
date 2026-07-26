package app

import (
	"context"
	"sync"

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

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Run(); err != nil {
			panic(err)
		}
	}()

	q := queue.New(ctx)
	workerErrCh := make(chan error)
	for i := 1; i <= a.cfg.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			wg.Done()
			w := worker.New(id)
			w.Start(ctx, q, workerErrCh)
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		handlerWorkerError(ctx, workerErrCh)
	}()

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
			// TODO: logging error
		}
	}
}

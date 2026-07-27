package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dimakropachev/image_resizer_service/internal/models"
	"github.com/dimakropachev/image_resizer_service/internal/queue"
	"github.com/kovidgoyal/imaging"
)

type Worker struct {
	ID   int
	stat *Statistics
	out  string
}

type Statistics struct {
	Total   int
	Fail    int
	Success int
	AvgTime time.Duration
	allTime time.Duration
}

func New(id int, out string) *Worker {
	return &Worker{
		ID:   id,
		stat: &Statistics{},
		out:  out,
	}
}

func (w *Worker) Start(ctx context.Context, q *queue.Queue, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for {
				task, ok := q.Get()
				if !ok {
					return
				}
				slog.Info("worker get task", slog.Int("worker_id", w.ID), slog.String("task_id", task.ID))

				task.Status = models.StatusProcessing

				w.stat.Total++
				src, err := imaging.Open(task.ImgPath)
				if err != nil {
					task.Status = models.StatusFailed
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't decode image: %w", task.ID, err)
				}

				now := time.Now()
				thumbnail := imaging.Fit(src, 150, 150, imaging.Lanczos)
				medium := imaging.Fit(src, 800, 600, imaging.Lanczos)
				dur := time.Since(now)

				w.stat.allTime += dur

				path := fmt.Sprintf("%s/%s/", w.out, task.ID)
				if err := os.MkdirAll(path, 0755); err != nil {
					task.Status = models.StatusFailed
					w.stat.Fail++
					errCh <- fmt.Errorf("couldn't create dir for processed img: %w", err)
					continue
				}
				fmt.Println(path)

				err = imaging.Save(thumbnail, path+"thumbnail.jpg")
				if err != nil {
					task.Status = models.StatusFailed
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't process thumbnail photo: %w", task.ID, err)
				}

				err = imaging.Save(medium, path+"medium.jpg")
				if err != nil {
					task.Status = models.StatusFailed
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't process medium photo: %w", task.ID, err)
				}
				task.Status = models.StatusDone
				w.stat.Success++

				slog.Info("worker processed task", slog.Int("worker_id", w.ID), slog.String("task_id", task.ID))
			}
		}
	}
}

func (w *Worker) GetStat() *Statistics {
	w.stat.AvgTime = w.stat.allTime / time.Duration(w.stat.Total)
	return w.stat
}

package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dimakropachev/image_resizer_service/internal/config"
	"github.com/dimakropachev/image_resizer_service/internal/models"
	"github.com/dimakropachev/image_resizer_service/internal/queue"
	"github.com/kovidgoyal/imaging"
)

type Worker struct {
	ID    int
	stat  *Statistics
	sizes []config.Size
}

type Statistics struct {
	Total   int           `json:"total_img"`
	Fail    int           `json:"fail_img"`
	Success int           `json:"success_img"`
	AvgTime time.Duration `json:"avg_time_ms"`
	allTime time.Duration
}

func New(id int, sizes []config.Size) *Worker {
	return &Worker{
		ID:    id,
		stat:  &Statistics{},
		sizes: sizes,
	}
}

func (w *Worker) Start(ctx context.Context, q *queue.Queue, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopped",
				slog.Int("id", w.ID),
			)
			return
		default:
			for {
				task, ok := q.Get()
				if !ok {
					return
				}
				slog.Info("worker get task",
					slog.Int("worker_id", w.ID),
					slog.String("task_id", task.ID),
				)

				task.Status = models.StatusProcessing

				w.stat.Total++
				src, err := imaging.Open(task.ImgPath)
				if err != nil {
					task.Status = models.StatusFailed
					task.Err = fmt.Errorf("couldn't decode image: %w", err)
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't decode image: %w", task.ID, err)
				}

				if err := os.MkdirAll(task.OutPath, 0755); err != nil {
					task.Status = models.StatusFailed
					task.Err = fmt.Errorf("couldn't create dir for processed img: %w", err)
					w.stat.Fail++
					errCh <- fmt.Errorf("couldn't create dir for processed img: %w", err)
					continue
				}

				now := time.Now()

				for _, size := range w.sizes {
					img := imaging.Fit(src, size.Width, size.Height, imaging.Lanczos)
					path := fmt.Sprintf("%s%s.jpg", task.OutPath, size.Name)
					err = imaging.Save(img, path)
					if err != nil {
						task.Status = models.StatusFailed
						task.Err = fmt.Errorf("couldn't save %s photo: %w", size.Name, err)
						w.stat.Fail++
						errCh <- fmt.Errorf("[%s] couldn't save %s photo: %w", task.ID, size.Name, err)
					}
				}

				dur := time.Since(now)

				w.stat.allTime += dur
				task.Status = models.StatusDone
				w.stat.Success++

				slog.Info("worker processed task",
					slog.Int("worker_id", w.ID),
					slog.String("task_id", task.ID),
				)
			}
		}
	}
}

func (w *Worker) GetStat() *Statistics {
	if w.stat.Total != 0 {
		w.stat.AvgTime = time.Duration((w.stat.allTime / time.Duration(w.stat.Total)).Milliseconds())
	}
	return w.stat
}

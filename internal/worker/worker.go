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
}

type Statistics struct {
	Total   int           `json:"total_img"`
	Fail    int           `json:"fail_img"`
	Success int           `json:"success_img"`
	AvgTime time.Duration `json:"avg_time_ms"`
	allTime time.Duration
}

func New(id int) *Worker {
	return &Worker{
		ID:   id,
		stat: &Statistics{},
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

				now := time.Now()
				thumbnail := imaging.Fit(src, 150, 150, imaging.Lanczos)
				medium := imaging.Fit(src, 800, 600, imaging.Lanczos)
				dur := time.Since(now)

				w.stat.allTime += dur

				if err := os.MkdirAll(task.OutPath, 0755); err != nil {
					task.Status = models.StatusFailed
					task.Err = fmt.Errorf("couldn't create dir for processed img: %w", err)
					w.stat.Fail++
					errCh <- fmt.Errorf("couldn't create dir for processed img: %w", err)
					continue
				}

				err = imaging.Save(thumbnail, task.OutPath+"thumbnail.jpg")
				if err != nil {
					task.Status = models.StatusFailed
					task.Err = fmt.Errorf("couldn't process thumbnail photo: %w", err)
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't process thumbnail photo: %w", task.ID, err)
				}

				err = imaging.Save(medium, task.OutPath+"medium.jpg")
				if err != nil {
					task.Status = models.StatusFailed
					task.Err = fmt.Errorf("couldn't process medium photo: %w", err)
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't process medium photo: %w", task.ID, err)
				}
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

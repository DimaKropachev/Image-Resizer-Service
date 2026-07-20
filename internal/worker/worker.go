package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/dimakropachev/image_resizer_service/internal/models"
	"github.com/kovidgoyal/imaging"
)

type Worker struct {
	ID   int
	stat *Statistics
}

type Statistics struct {
	Total   int
	Fail    int
	Success int
	AvgTime time.Duration
	allTime time.Duration
}

func New(id int) *Worker {
	return &Worker{
		ID:   id,
		stat: &Statistics{},
	}
}

func (w *Worker) Start(ctx context.Context, queue chan *models.Task, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for task := range queue {
				w.stat.Total++
				src, err := imaging.Open(task.ImgPath)
				if err != nil {
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't decode image: %w", task.ID, err)
				}

				now := time.Now()
				thumbnail := imaging.Fit(src, 150, 150, imaging.Lanczos)
				medium := imaging.Fit(src, 800, 600, imaging.Lanczos)
				dur := time.Since(now)

				w.stat.allTime += dur

				path := fmt.Sprintf("./storage/images/processed/%s/", task.ID)

				err = imaging.Save(thumbnail, path+"thumbnail.jpg")
				if err != nil {
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't process thumbnail photo: %w", task.ID, err)
				}

				err = imaging.Save(medium, path+"medium.jpg")
				if err != nil {
					w.stat.Fail++
					errCh <- fmt.Errorf("[%s] couldn't process medium photo: %w", task.ID, err)
				}
				w.stat.Success++
			}
		}
	}
}

func (w *Worker) GetStat() *Statistics {
	w.stat.AvgTime = w.stat.allTime / time.Duration(w.stat.Total)
	return w.stat
}

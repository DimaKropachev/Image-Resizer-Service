package queue

import (
	"context"
	"fmt"

	"github.com/dimakropachev/image_resizer_service/internal/models"
)

type Queue struct {
	ctx context.Context
	q   chan *models.Task
}

func New(ctx context.Context) *Queue {
	return &Queue{
		ctx: ctx,
		q:   make(chan *models.Task),
	}
}

func (q *Queue) Add(task *models.Task) {
	for {
		select {
		case q.q <- task:
			return
		case <-q.ctx.Done():
			return
		}
	}
}

func (q *Queue) Get() (*models.Task, bool) {
	for {
		select {
		case task, ok := <-q.q:
			return task, ok
		case <-q.ctx.Done():
			return nil, false
		}
	}
}

func (q *Queue) Close() {
	fmt.Println("close")
	close(q.q)
}

package service

import (
	"context"

	"github.com/dimakropachev/image_resizer_service/internal/models"
	"github.com/dimakropachev/image_resizer_service/internal/queue"
)

type Repository interface {
	AddTask(context.Context, *models.Task) error
	GetStatus(context.Context, string) (string, error, error)
	DeleteTask(context.Context, string) error
}

type Service struct {
	repo Repository
	q    *queue.Queue
}

func New(repo Repository, q *queue.Queue) *Service {
	return &Service{
		repo: repo,
		q:    q,
	}
}

func (s *Service) AddTask(ctx context.Context, task *models.Task) error {
	s.q.Add(task)
	return s.repo.AddTask(ctx, task)
}

func (s *Service) GetStatus(ctx context.Context, id string) (string, error, error) {
	return s.repo.GetStatus(ctx, id)
}

func (s *Service) DeleteTask(ctx context.Context, id string) error {
	return s.repo.DeleteTask(ctx, id)
}

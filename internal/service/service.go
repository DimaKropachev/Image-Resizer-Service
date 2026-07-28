package service

import (
	"context"
	"log/slog"

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
	if err := s.repo.AddTask(ctx, task); err != nil {
		slog.Error("failed to add task",
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (s *Service) GetStatus(ctx context.Context, id string) (string, error, error) {
	status, tErr, err := s.repo.GetStatus(ctx, id)
	if tErr != nil {
		slog.Debug("error during task execution",
			slog.String("task_id", id),
			slog.String("error", err.Error()),
		)
	}
	if err != nil {
		slog.Error("failed to get status task",
			slog.String("error", err.Error()),
		)
		return "", nil, err
	}
	return status, tErr, nil
}

func (s *Service) DeleteTask(ctx context.Context, id string) error {
	if err := s.repo.DeleteTask(ctx, id); err != nil {
		slog.Error("failed to delete task",
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

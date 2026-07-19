package service

import (
	"context"

	"github.com/dimakropachev/image_resizer_service/internal/models"
)

type Repository interface {
	AddTask(context.Context, *models.Task) error
	GetStatus(context.Context, string) (string, error, error)
	DeleteTask(context.Context, string) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) AddTask(ctx context.Context, task *models.Task) error {
	return s.repo.AddTask(ctx, task)
}


func (s *Service) GetStatus(ctx context.Context, id string) (string, error, error) {
	return s.repo.GetStatus(ctx, id)
}

func (s *Service) DeleteTask(ctx context.Context, id string) error {
	return s.repo.DeleteTask(ctx, id)
}

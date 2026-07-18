package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/dimakropachev/image_resizer_service/internal/models"
)

type Storage struct {
	mu    *sync.Mutex
	tasks map[string]*models.Task
}

func New() *Storage {
	return &Storage{
		mu:    &sync.Mutex{},
		tasks: make(map[string]*models.Task),
	}
}

func (s *Storage) AddTask(ctx context.Context, task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.tasks[task.ID]
	if exists {
		return fmt.Errorf("task with ID: %s is exists", task.ID)
	}

	s.tasks[task.ID] = task
	return nil
}

// return status, error task if there is one, error
func (s *Storage) GetStatus(ctx context.Context, id string) (string, error, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.tasks[id]
	if !exists {
		return "", nil, fmt.Errorf("task with ID: %s not found", id)
	}

	return s.tasks[id].Status, s.tasks[id].Err, nil
}

func (s *Storage) DeleteTask(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task with ID: %s not found", id)
	}

	delete(s.tasks, id)
	return nil
}

func (s *Storage) DeleteAllTask(ctx context.Context) {}
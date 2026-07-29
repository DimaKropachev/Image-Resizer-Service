package wm

import (
	"fmt"
	"sync"

	"github.com/dimakropachev/image_resizer_service/internal/worker"
)

type WorkerManager struct {
	mu      sync.RWMutex
	workers map[int]*worker.Worker
}

func New() *WorkerManager {
	return &WorkerManager{
		mu:      sync.RWMutex{},
		workers: make(map[int]*worker.Worker),
	}
}

func (wm *WorkerManager) Register(id int, w *worker.Worker) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.workers[id] = w
}

func (wm *WorkerManager) Unregister(id int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	delete(wm.workers, id)
}

func (wm *WorkerManager) GetStat(id int) (*worker.Statistics, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	w, exists := wm.workers[id]
	if !exists {
		return nil, fmt.Errorf("worker not found")
	}

	return w.GetStat(), nil
}

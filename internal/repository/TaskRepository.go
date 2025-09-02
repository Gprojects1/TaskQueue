package repository

import (
	"fmt"
	"sync"

	"TaskQueue/internal/model"
)

type TaskRepository interface {
	Create(task *model.Task) error
	GetByID(id string) (*model.Task, bool)
	Update(task *model.Task) error
	Exists(id string) bool
	GetAll() map[string]*model.Task
}

type inMemoryTaskRepository struct {
	tasks map[string]*model.Task
	mu    sync.RWMutex
}

func NewInMemoryTaskRepository() TaskRepository {
	return &inMemoryTaskRepository{
		tasks: make(map[string]*model.Task),
	}
}

func (r *inMemoryTaskRepository) Create(task *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; exists {
		return fmt.Errorf("task with id %s already exists", task.ID)
	}

	r.tasks[task.ID] = task
	return nil
}

func (r *inMemoryTaskRepository) GetByID(id string) (*model.Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	return task, exists
}

func (r *inMemoryTaskRepository) Update(task *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return fmt.Errorf("task with id %s not found", task.ID)
	}

	r.tasks[task.ID] = task
	return nil
}

func (r *inMemoryTaskRepository) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.tasks[id]
	return exists
}

func (r *inMemoryTaskRepository) GetAll() map[string]*model.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Возвращаем копию
	result := make(map[string]*model.Task)
	for id, task := range r.tasks {
		result[id] = task
	}
	return result
}

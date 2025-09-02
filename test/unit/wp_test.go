package unit

import (
	"TaskQueue/internal/model"
	"TaskQueue/internal/repository"
	"TaskQueue/queue"
	"testing"
)

func TestWorkerPool_Enqueue(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	pool := queue.NewWorkerPool(2, 5, repo)

	task := &model.Task{
		ID:         "test1",
		Payload:    "test",
		MaxRetries: 3,
	}

	err := pool.Enqueue(task)
	if err != nil {
		t.Errorf("Failed to enqueue task: %v", err)
	}
}

func TestWorkerPool_QueueFull(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	pool := queue.NewWorkerPool(1, 1, repo)

	// Fill the queue
	task1 := &model.Task{
		ID:         "test1",
		Payload:    "test",
		MaxRetries: 3,
	}
	pool.Enqueue(task1)

	// Try to add another task (should fail)
	task2 := &model.Task{
		ID:         "test2",
		Payload:    "test",
		MaxRetries: 3,
	}
	err := pool.Enqueue(task2)

	if err == nil {
		t.Error("Expected error when queue is full")
	}
}

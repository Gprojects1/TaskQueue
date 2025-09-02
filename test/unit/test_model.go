package unit

import (
	"TaskQueue/internal/model"
	"testing"
	"time"
)

func TestModel_Task_SetStatus(t *testing.T) {
	task := &model.Task{
		ID:     "test1",
		Status: "queued",
	}

	task.SetStatus("running")
	if task.GetStatus() != "running" {
		t.Errorf("Expected status 'running', got '%s'", task.GetStatus())
	}
}

func TestModel_Task_IncrementRetries(t *testing.T) {
	task := &model.Task{
		ID:         "test1",
		MaxRetries: 3,
		Retries:    0,
	}

	retries := task.IncrementRetries()
	if retries != 1 {
		t.Errorf("Expected 1 retry, got %d", retries)
	}
}

func TestModel_Task_ConcurrentAccess(t *testing.T) {
	task := &model.Task{
		ID:     "test1",
		Status: "queued",
	}

	for i := 0; i < 100; i++ {
		go func() {
			task.SetStatus("running")
			task.IncrementRetries()
		}()
	}

	time.Sleep(100 * time.Millisecond)
}

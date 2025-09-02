package unit

import (
	"TaskQueue/internal/model"
	"TaskQueue/internal/repository"
	"testing"
)

func TestRepository_InMemoryTaskRepository(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	task := &model.Task{
		ID:         "test1",
		Payload:    "test payload",
		MaxRetries: 3,
	}

	// Test Create
	err := repo.Create(task)
	if err != nil {
		t.Errorf("Failed to create task: %v", err)
	}

	// Test duplicate
	err = repo.Create(task)
	if err == nil {
		t.Error("Expected error when creating duplicate task")
	}

	// Test GetByID
	retrievedTask, exists := repo.GetByID("test1")
	if !exists {
		t.Error("Task should exist")
	}
	if retrievedTask.ID != "test1" {
		t.Errorf("Expected task ID 'test1', got '%s'", retrievedTask.ID)
	}
}

func TestRepository_UpdateNonExistentTask(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	task := &model.Task{ID: "nonexistent"}

	err := repo.Update(task)
	if err == nil {
		t.Error("Expected error when updating non-existent task")
	}
}

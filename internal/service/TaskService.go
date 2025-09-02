package service

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"TaskQueue/internal/model"
	"TaskQueue/internal/repository"
	"TaskQueue/queue"
)

type QueueService interface {
	Enqueue(task *model.Task) error
	GetTaskStatus(id string) (string, bool)
	StartWorkers()
	Shutdown()
}

type queueService struct {
	taskRepo   repository.TaskRepository
	workerPool queue.WorkerPool
	workers    int
	queueSize  int
}

func NewQueueService(taskRepo repository.TaskRepository, workers, queueSize int) QueueService {
	workerPool := queue.NewWorkerPool(workers, queueSize, taskRepo)
	return &queueService{
		taskRepo:   taskRepo,
		workerPool: workerPool,
		workers:    workers,
		queueSize:  queueSize,
	}
}

func (s *queueService) Enqueue(task *model.Task) error {
	if s.taskRepo.Exists(task.ID) {
		return fmt.Errorf("task with id %s already exists", task.ID)
	}

	task.SetStatus("queued")

	if err := s.taskRepo.Create(task); err != nil {
		return err
	}

	if err := s.workerPool.Enqueue(task); err != nil {
		task.SetStatus("failed")
		s.taskRepo.Update(task)
		return fmt.Errorf("failed to enqueue task: %v", err)
	}

	return nil
}

func (s *queueService) GetTaskStatus(id string) (string, bool) {
	task, exists := s.taskRepo.GetByID(id)
	if !exists {
		return "", false
	}
	return task.GetStatus(), true
}

func (s *queueService) StartWorkers() {
	s.workerPool.Start()
}

func (s *queueService) Shutdown() {
	s.workerPool.Shutdown()
}

func processTask(task *model.Task) {
	processingTime := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(processingTime)

	if rand.Float64() < 0.2 {
		retries := task.IncrementRetries()

		if retries >= task.MaxRetries {
			task.SetStatus("failed")
			log.Printf("Task %s failed after %d retries", task.ID, retries)
		} else {
			task.SetStatus("queued")
			log.Printf("Task %s failed, retry %d/%d", task.ID, retries, task.MaxRetries)
		}
	} else {
		task.SetStatus("done")
		log.Printf("Task %s completed successfully", task.ID)
	}
}

package queue

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"TaskQueue/internal/model"
	"TaskQueue/internal/repository"
)

type WorkerPool interface {
	Enqueue(task *model.Task) error
	Start()
	Shutdown()
}

type workerPool struct {
	tasks    chan *model.Task
	workers  int
	shutdown chan struct{}
	wg       sync.WaitGroup
	taskRepo repository.TaskRepository
}

func NewWorkerPool(workers, queueSize int, taskRepo repository.TaskRepository) WorkerPool {
	return &workerPool{
		tasks:    make(chan *model.Task, queueSize),
		workers:  workers,
		shutdown: make(chan struct{}),
		taskRepo: taskRepo,
	}
}

func (wp *workerPool) Enqueue(task *model.Task) error {
	select {
	case wp.tasks <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

func (wp *workerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *workerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.shutdown:
			return
		case task := <-wp.tasks:
			wp.processTask(task, id)
		}
	}
}

func (wp *workerPool) processTask(task *model.Task, workerID int) {
	// Обновляем статус на "running"
	task.SetStatus("running")
	wp.taskRepo.Update(task)

	processingTime := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(processingTime)

	// 20% вероятность ошибки
	if rand.Float64() < 0.2 {
		retries := task.IncrementRetries()

		if retries >= task.MaxRetries {
			task.SetStatus("failed")
			wp.taskRepo.Update(task)
			log.Printf("Worker %d: Task %s failed after %d retries", workerID, task.ID, retries)
		} else {
			task.SetStatus("queued")
			wp.taskRepo.Update(task)
			//бэкофф
			backoff := time.Duration(1<<uint(retries)) * time.Second
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			retryDelay := backoff + jitter

			log.Printf("Worker %d: Task %s failed, retry %d/%d in %v",
				workerID, task.ID, retries, task.MaxRetries, retryDelay)

			// Перезапускаем задачу после задержки
			time.AfterFunc(retryDelay, func() {
				select {
				case wp.tasks <- task:
				case <-wp.shutdown:
				}
			})
		}
	} else {
		task.SetStatus("done")
		wp.taskRepo.Update(task)
		log.Printf("Worker %d: Task %s completed successfully", workerID, task.ID)
	}
}

func (wp *workerPool) Shutdown() {
	close(wp.shutdown)
	wp.wg.Wait()
}

var ErrQueueFull = &QueueError{Message: "queue is full"}

type QueueError struct {
	Message string
}

func (e *QueueError) Error() string {
	return e.Message
}

package model

import "sync"

type Task struct {
	ID         string `json:"id"`
	Payload    string `json:"payload"`
	MaxRetries int    `json:"max_retries"`
	Retries    int    `json:"-"`
	Status     string `json:"status"`
	mu         sync.Mutex
}

func (t *Task) SetStatus(status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
}

func (t *Task) IncrementRetries() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Retries++
	return t.Retries
}

func (t *Task) GetStatus() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Status
}

func (t *Task) GetRetries() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Retries
}

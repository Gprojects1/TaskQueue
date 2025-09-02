package main

import (
	"TaskQueue/internal/controller"
	"TaskQueue/internal/repository"
	"TaskQueue/internal/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegration_CompleteFlow(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	queueService := service.NewQueueService(repo, 2, 5)
	httpController := controller.NewHTTPController(queueService)

	queueService.StartWorkers()
	defer queueService.Shutdown()

	task := map[string]interface{}{
		"id":          "integration-test",
		"payload":     "integration data",
		"max_retries": 2,
	}
	body, _ := json.Marshal(task)

	req := httptest.NewRequest("POST", "/enqueue", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	httpController.EnqueueHandler(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	time.Sleep(600 * time.Millisecond)

	req = httptest.NewRequest("GET", "/status?id=integration-test", nil)
	w = httptest.NewRecorder()

	httpController.StatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var statusResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &statusResponse)

	if statusResponse["status"] != "done" && statusResponse["status"] != "failed" {
		t.Errorf("Expected status 'done' or 'failed', got '%s'", statusResponse["status"])
	}
}

func TestIntegration_HealthCheck(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	queueService := service.NewQueueService(repo, 2, 5)
	httpController := controller.NewHTTPController(queueService)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	httpController.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestIntegration_QueueFull(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()

	queueService := service.NewQueueService(repo, 0, 1)
	httpController := controller.NewHTTPController(queueService)

	task1 := map[string]interface{}{
		"id":          "test1",
		"payload":     "data1",
		"max_retries": 1,
	}
	body1, _ := json.Marshal(task1)

	req := httptest.NewRequest("POST", "/enqueue", bytes.NewReader(body1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	httpController.EnqueueHandler(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("First task should be accepted, got %d", w.Code)
	}

	time.Sleep(10 * time.Millisecond)

	task2 := map[string]interface{}{
		"id":          "test2",
		"payload":     "data2",
		"max_retries": 1,
	}
	body2, _ := json.Marshal(task2)

	req = httptest.NewRequest("POST", "/enqueue", bytes.NewReader(body2))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	httpController.EnqueueHandler(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 for full queue, got %d. Response: %s", w.Code, w.Body.String())
	}
}

func TestIntegration_RetryMechanism(t *testing.T) {
	repo := repository.NewInMemoryTaskRepository()
	queueService := service.NewQueueService(repo, 1, 5)
	httpController := controller.NewHTTPController(queueService)

	queueService.StartWorkers()
	defer queueService.Shutdown()

	task := map[string]interface{}{
		"id":          "retry-test",
		"payload":     "retry data",
		"max_retries": 3,
	}
	body, _ := json.Marshal(task)

	req := httptest.NewRequest("POST", "/enqueue", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	httpController.EnqueueHandler(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	time.Sleep(2 * time.Second)

	req = httptest.NewRequest("GET", "/status?id=retry-test", nil)
	w = httptest.NewRecorder()

	httpController.StatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var statusResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &statusResponse)

	if statusResponse["status"] != "done" && statusResponse["status"] != "failed" {
		t.Errorf("Expected status 'done' or 'failed', got '%s'", statusResponse["status"])
	}
}

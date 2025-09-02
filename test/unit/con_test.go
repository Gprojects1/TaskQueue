package unit

import (
	"TaskQueue/internal/controller"
	"TaskQueue/internal/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockQueueService struct {
	enqueueErr error
	status     string
	exists     bool
}

func (m *MockQueueService) Enqueue(task *model.Task) error {
	return m.enqueueErr
}

func (m *MockQueueService) GetTaskStatus(id string) (string, bool) {
	return m.status, m.exists
}

func (m *MockQueueService) StartWorkers() {}
func (m *MockQueueService) Shutdown()     {}

func TestController_EnqueueHandler_Success(t *testing.T) {
	mockService := &MockQueueService{}
	controller := controller.NewHTTPController(mockService)

	task := map[string]interface{}{
		"id":          "test1",
		"payload":     "test data",
		"max_retries": 3,
	}
	body, _ := json.Marshal(task)

	req := httptest.NewRequest("POST", "/enqueue", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.EnqueueHandler(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "accepted" {
		t.Errorf("Expected status 'accepted', got '%s'", response["status"])
	}
}

func TestController_EnqueueHandler_InvalidJSON(t *testing.T) {
	controller := controller.NewHTTPController(&MockQueueService{})

	req := httptest.NewRequest("POST", "/enqueue", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.EnqueueHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
	}
}

func TestController_EnqueueHandler_MissingFields(t *testing.T) {
	controller := controller.NewHTTPController(&MockQueueService{})

	invalidTask := map[string]interface{}{
		"id": "test1",
		// missing payload and max_retries
	}
	body, _ := json.Marshal(invalidTask)

	req := httptest.NewRequest("POST", "/enqueue", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	controller.EnqueueHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing fields, got %d", w.Code)
	}
}

func TestController_HealthHandler(t *testing.T) {
	controller := controller.NewHTTPController(&MockQueueService{})

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	controller.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestController_StatusHandler(t *testing.T) {
	mockService := &MockQueueService{
		status: "running",
		exists: true,
	}
	controller := controller.NewHTTPController(mockService)

	req := httptest.NewRequest("GET", "/status?id=test1", nil)
	w := httptest.NewRecorder()

	controller.StatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "running" {
		t.Errorf("Expected status 'running', got '%s'", response["status"])
	}
}

func TestController_StatusHandler_MissingID(t *testing.T) {
	controller := controller.NewHTTPController(&MockQueueService{})

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	controller.StatusHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing id, got %d", w.Code)
	}
}

func TestController_StatusHandler_NotFound(t *testing.T) {
	mockService := &MockQueueService{exists: false}
	controller := controller.NewHTTPController(mockService)

	req := httptest.NewRequest("GET", "/status?id=nonexistent", nil)
	w := httptest.NewRecorder()

	controller.StatusHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent task, got %d", w.Code)
	}
}

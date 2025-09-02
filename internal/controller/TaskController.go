package controller

import (
	"encoding/json"
	"net/http"

	"TaskQueue/internal/model"
	"TaskQueue/internal/service"
)

type HTTPController struct {
	queueService service.QueueService
}

func NewHTTPController(queueService service.QueueService) *HTTPController {
	return &HTTPController{
		queueService: queueService,
	}
}

func (c *HTTPController) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task model.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" || task.Payload == "" || task.MaxRetries <= 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if err := c.queueService.Enqueue(&task); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
		"id":     task.ID,
	})
}

func (c *HTTPController) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (c *HTTPController) StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	status, exists := c.queueService.GetTaskStatus(id)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"status": status,
	})
}

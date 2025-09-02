package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"TaskQueue/config"
	"TaskQueue/internal/controller"
	"TaskQueue/internal/repository"
	"TaskQueue/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("Starting with %d workers, queue size %d, port %s",
		cfg.Workers, cfg.QueueSize, cfg.Port)

	taskRepo := repository.NewInMemoryTaskRepository()
	queueService := service.NewQueueService(taskRepo, cfg.Workers, cfg.QueueSize)
	httpController := controller.NewHTTPController(queueService)

	queueService.StartWorkers()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /enqueue", httpController.EnqueueHandler)
	mux.HandleFunc("GET /healthz", httpController.HealthHandler)
	mux.HandleFunc("GET /status", httpController.StatusHandler)

	server := &http.Server{
		Addr:    "127.0.0.1:9000",
		Handler: mux,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	log.Println("Initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	queueService.Shutdown()

	log.Println("Shutdown completed")
}

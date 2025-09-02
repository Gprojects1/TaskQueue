package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	Workers   int
	QueueSize int
	Port      string
}

func LoadConfig() Config {
	workers := getEnvInt("WORKERS", 4)
	queueSize := getEnvInt("QUEUE_SIZE", 64)
	port := getEnvString("PORT", "8080")

	return Config{
		Workers:   workers,
		QueueSize: queueSize,
		Port:      port,
	}
}

func getEnvInt(key string, defaultValue int) int {
	if env := os.Getenv(key); env != "" {
		var value int
		if n, err := fmt.Sscanf(env, "%d", &value); err == nil && n != 1 {
			log.Printf("Invalid %s value, using default: %d", key, defaultValue)
		} else if err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvString(key, defaultValue string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return defaultValue
}

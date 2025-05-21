package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Server   ServerConfig
	RabbitMQ RabbitMQConfig
	Storage  StorageConfig
}

type ServerConfig struct {
	Port string
}

type RabbitMQConfig struct {
	URL          string
	QueueName    string
	ExchangeName string
	RoutingKey   string
}

type StorageConfig struct {
	Type      string // "local", "s3", etc.
	LocalPath string
}

// NewConfig creates a new configuration with default values
// We can add another service and change it anytime
func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			QueueName:    getEnv("RABBITMQ_QUEUE", "image_processing"),
			ExchangeName: getEnv("RABBITMQ_EXCHANGE", "image_exchange"),
			RoutingKey:   getEnv("RABBITMQ_ROUTING_KEY", "image_key"),
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", filepath.Join(".", "storage")),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

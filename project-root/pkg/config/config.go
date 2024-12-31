package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"project/pkg/models"
)

func LoadConfig() (*models.Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	configs := &models.Config{
		REDIS_URL1:            os.Getenv("REDIS_URL1"),
		REDIS_URL2:            os.Getenv("REDIS_URL2"),
		RABBITMQ_URL:          os.Getenv("RABBITMQ_URL"),
		SUBSCRIBER_QUEUE:      os.Getenv("SUBSCRIBER_QUEUE"),
		PUBLISHER_EXCHANGE:    os.Getenv("PUBLISHER_EXCHANGE"),
		PUBLISHER_ROUTING_KEY: os.Getenv("PUBLISHER_ROUTING_KEY"),
	}

	if configs.REDIS_URL1 == "" || configs.REDIS_URL2 == "" || configs.RABBITMQ_URL == "" || configs.SUBSCRIBER_QUEUE == "" || configs.PUBLISHER_EXCHANGE == "" || configs.PUBLISHER_ROUTING_KEY == "" {
		return nil, fmt.Errorf("missing one or more required environment variables")
	}

	return configs, nil
}

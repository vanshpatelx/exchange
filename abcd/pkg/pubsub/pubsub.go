package pubsub

import (
	"github.com/streadway/amqp"
	"log"
	"sync"
)

var (
	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
	onceRabbit    sync.Once
)

// GetRabbitChannel returns the RabbitMQ channel instance (singleton).
func GetRabbitChannel() *amqp.Channel {
	onceRabbit.Do(func() {
		// Initialize RabbitMQ connection and channel only once
		var err error
		rabbitConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			log.Fatalf("Failed to connect to RabbitMQ: %v", err)
		}

		rabbitChannel, err = rabbitConn.Channel()
		if err != nil {
			log.Fatalf("Failed to open a channel: %v", err)
		}
	})
	return rabbitChannel
}

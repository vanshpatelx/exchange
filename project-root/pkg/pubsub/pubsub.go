package pubsub

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type PubSub struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPubSub(rabbitMQURL string) (*PubSub, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &PubSub{
		conn:    conn,
		channel: channel,
	}, nil
}

func (ps *PubSub) Publish(exchange, routingKey, message string) error {
	err := ps.channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	log.Printf("Published message: %s to exchange: %s with routing key: %s", message, exchange, routingKey)
	return nil
}

func (ps *PubSub) Subscribe(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := ps.channel.Consume(
		queueName,
		"",        // consumer tag
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}
	return msgs, nil
}
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"project/pkg/cache"
	"project/pkg/exchange"
	"project/pkg/models"
	"project/pkg/pubsub"
	"project/pkg/config"
	"syscall"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cacheInstance1 := cache.NewCache(config.REDIS_URL1)
	cacheInstance2 := cache.NewCache(config.REDIS_URL2)

	pubsubInstance, err := pubsub.NewPubSub(config.RABBITMQ_URL)
	if err != nil {
		log.Fatalf("Failed to create PubSub instance: %v", err)
	}

	exchangeInstance := exchange.NewExchange(pubsubInstance, cacheInstance1, cacheInstance2, config)

	go subscribeToAddOrderMatchQueue(pubsubInstance, exchangeInstance, config)

	gracefulShutdown()
}


func subscribeToAddOrderMatchQueue(pubsubInstance *pubsub.PubSub, exchangeInstance *exchange.Exchange, config *models.Config) {
	msgs, err := pubsubInstance.Subscribe(config.SUBSCRIBER_QUEUE)
	if err != nil {
		log.Fatalf("Failed to subscribe to addOrderMatch queue: %v", err)
	}

	// Process messages concurrently
	for msg := range msgs {
		go processOrderMessage(msg.Body, exchangeInstance)
	}
}

func processOrderMessage(msgBody []byte, exchangeInstance *exchange.Exchange) {
	var exchangeMsg models.ExchangeMsg
	err := json.Unmarshal(msgBody, &exchangeMsg)
	if err != nil {
		log.Printf("Error unmarshalling order message: %v", err)
		return
	}

	// Handle different tas
	switch exchangeMsg.Task {
	case 0:
		// Place an order
		exchangeInstance.PlaceOrder(&exchangeMsg.Order)
	case 1:
		// Handle settlement
		log.Printf("Settlement task received for order ID: %d", exchangeMsg.Order.Id)
	default:
		log.Printf("Unknown task type: %d", exchangeMsg.Task)
	}

	fmt.Printf("Processed order for stock: %d, User: %d, Quantity: %d, Type: %d\n", exchangeMsg.Order.Ticker, exchangeMsg.Order.User, exchangeMsg.Order.Quantity, exchangeMsg.Order.Type)
}


func gracefulShutdown() {
	done := make(chan bool)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Received termination signal. Shutting down...")
		done <- true 
	}()

	<-done
}

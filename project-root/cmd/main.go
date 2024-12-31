package main

import (
	"encoding/json"
	"fmt"
	"log"
	"project/pkg/cache"
	"project/pkg/exchange"
	"project/pkg/models"
	"project/pkg/pubsub"
)

func main() {
	redisURL := "localhost:6379"
	cacheInstance := cache.NewCache(redisURL)

	rabbitMQURL := "amqp://guest:guest@localhost:5672/"
	pubsubInstance, err := pubsub.NewPubSub(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to create PubSub instance: %v", err)
	}

	exchangeInstance := exchange.NewExchange(redisURL, pubsubInstance, cacheInstance)

	go subscribeToAddOrderMatchQueue(pubsubInstance, exchangeInstance)

	select {}
}

type Order struct {
	Id       int     `json:"id"`
	StockId  int     `json:"stock_id"`
	Type     string  `json:"type"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	UserId   int     `json:"user_id"`
}

func subscribeToAddOrderMatchQueue(pubsubInstance *pubsub.PubSub, exchangeInstance *exchange.Exchange) {
	msgs, err := pubsubInstance.Subscribe("addOrderMatch")
	if err != nil {
		log.Fatalf("Failed to subscribe to addOrderMatch queue: %v", err)
	}

	for msg := range msgs {
		processOrderMessage(msg.Body, exchangeInstance)
	}
}

func processOrderMessage(msgBody []byte, exchangeInstance *exchange.Exchange) {
	var exchangeMsg models.ExchangeMsg
	err := json.Unmarshal(msgBody, &exchangeMsg)
	if err != nil {
		log.Printf("Error unmarshalling order message: %v", err)
		return
	}

	switch exchangeMsg.Task {
	case 0: exchangeInstance.PlaceOrder(&exchangeMsg.Order) //order
	case 1: // settlement
	}

	fmt.Printf("Processed order for stock: %d, User: %d, Quantity: %d, Type: %d\n", exchangeMsg.Order.Ticker, exchangeMsg.Order.User, exchangeMsg.Order.Quantity, exchangeMsg.Order.Type)
}

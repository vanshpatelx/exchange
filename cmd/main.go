package main

import (
	"fmt"
	"exchange/pkg/orderbook"
	"math/rand"
	"time"
)

func main() {
	// Create a new order book
	orderBook := orderbook.NewOrderBook()

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate 100 orders
	for i := 1; i <= 100; i++ {
		// Randomly choose order type (LIMIT or MARKET)
		orderType := rand.Intn(2) == 0 // 0 for LIMIT, 1 for MARKET

		// Randomly choose side (BUY or SELL)
		side := rand.Intn(2) == 0 // 0 for BUY, 1 for SELL

		// Randomly set price for LIMIT orders, or leave it out for MARKET orders
		price := uint32(0)
		if orderType {
			price = uint32(rand.Intn(200) + 1) // Random price between 1 and 200 for LIMIT orders
		}

		// Generate a random quantity between 1 and 50
		quantity := uint32(rand.Intn(50) + 1)

		// Create an order and add it to the order book
		order := &orderbook.Order{
			ID:       uint64(i),
			Type:     orderType,
			Side:     side,
			Price:    price,
			Quantity: quantity,
		}
		go orderBook.AddOrder(order)
	}
	
	time.Sleep(10 * time.Millisecond)

	fmt.Println("=== Order Book Statex===")

	fmt.Println("\nBuy Levels (Max Heap):")
	for _, priceLevel := range orderBook.BuyLevels().GetLevels() {
		fmt.Printf("Price: %d, Volume: %d\n", priceLevel.Price, priceLevel.Volume)
		for _, order := range priceLevel.Orders {
			fmt.Printf("\tOrder ID: %d, Quantity: %d\n", order.ID, order.Quantity)
		}
	}
	fmt.Println("\nSell Levels (Min Heap):")
	for _, priceLevel := range orderBook.SellLevels().GetLevels() {
		fmt.Printf("Price: %d, Volume: %d\n", priceLevel.Price, priceLevel.Volume)
		for _, order := range priceLevel.Orders {
			fmt.Printf("\tOrder ID: %d, Quantity: %d\n", order.ID, order.Quantity)
		}
	}

}


package main

import (
	"testing"
	"exchange/pkg/orderbook"
	"math/rand"
	"time"
)

func BenchmarkOrderBook_AddOrder(b *testing.B) {
    // Initialize the order book
    orderBook := orderbook.NewOrderBook()

    // Seed the random number generator
    rand.Seed(time.Now().UnixNano())

    b.ResetTimer() // Reset any time spent on setup

    // Benchmark adding orders
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            orderType := rand.Intn(2) == 0 // 0 for LIMIT, 1 for MARKET
            side := rand.Intn(2) == 0      // 0 for BUY, 1 for SELL
            price := uint32(0)
            if orderType {
                price = uint32(rand.Intn(200) + 1) // Random price for LIMIT orders
            }
            quantity := uint32(rand.Intn(50) + 1)

            order := &orderbook.Order{
                ID:       uint64(rand.Intn(1000000)), // Random order ID
                Type:     orderType,
                Side:     side,
                Price:    price,
                Quantity: quantity,
            }

            // Add the order to the order book
            orderBook.AddOrder(order)
        }
    })
}

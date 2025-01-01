package cache

import (
	"context"
	"sync"
	"testing"
)

func setupRedis() *Cache {
	redisURL := "localhost:6379" // Update if necessary
	cache := NewCache(redisURL)
	cache.GetRedisClient() // Initialize Redis client
	return cache
}

func teardownRedis(cache *Cache) {
	ctx := context.Background()
	client := cache.GetRedisClient()
	client.FlushAll(ctx) // Clear Redis for clean state
}

func TestCache_GetAndSetStock(t *testing.T) {
	cache := setupRedis()
	defer teardownRedis(cache)

	userID := 1
	tickerID := 1001
	quantity := 10
	price := 50

	// Ensure stock data is set before getting it
	success := cache.SetStock(userID, tickerID, quantity, price, false, false)
	if !success {
		t.Fatalf("Failed to set stock for userID %d", userID)
	}

	// Get stocks
	stocks, err := cache.GetStocks(userID)
	if err != nil {
		t.Fatalf("Failed to get stocks for userID %d: %v", userID, err)
	}

	if len(stocks) != 1 {
		t.Fatalf("Expected 1 stock, got %d", len(stocks))
	}

	stock := stocks[0]
	if stock.TickerID != tickerID || stock.Quantity != quantity || stock.Price != price {
		t.Errorf("Unexpected stock data: %+v", stock)
	}
}

func TestCache_ConcurrentStockUpdates(t *testing.T) {
	cache := setupRedis()
	defer teardownRedis(cache)

	userID := 1
	tickerID := 1001
	initialQuantity := 100
	initialPrice := 50

	// Ensure initial stock is set
	success := cache.SetStock(userID, tickerID, initialQuantity, initialPrice, false, false)
	if !success {
		t.Fatalf("Failed to set initial stock for userID %d", userID)
	}

	var wg sync.WaitGroup
	concurrentUpdates := 10

	for i := 0; i < concurrentUpdates; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.SetStock(userID, tickerID, id+1, initialPrice+id, false, false)
		}(i)
	}

	wg.Wait()

	// Verify updated stock
	stocks, err := cache.GetStocks(userID)
	if err != nil {
		t.Fatalf("Failed to get stocks: %v", err)
	}

	if len(stocks) != 1 {
		t.Fatalf("Expected 1 stock, got %d", len(stocks))
	}

	stock := stocks[0]
	expectedQuantity := initialQuantity + (concurrentUpdates * (concurrentUpdates + 1) / 2)
	if stock.Quantity != expectedQuantity {
		t.Errorf("Unexpected quantity: got %d, want %d", stock.Quantity, expectedQuantity)
	}
}

func TestCache_BalanceOperations(t *testing.T) {
	cache := setupRedis()
	defer teardownRedis(cache)

	userID := 1
	initialBalance := 1000
	amount := 200

	// Ensure initial balance is set before retrieval
	success := cache.SetBalance(userID, initialBalance, false, false)
	if !success {
		t.Fatalf("Failed to set initial balance for userID %d", userID)
	}

	// Update balance
	success = cache.SetBalance(userID, amount, false, false)
	if !success {
		t.Fatalf("Failed to update balance for userID %d", userID)
	}

	// Get balance
	balance, err := cache.GetBalance(userID)
	if err != nil {
		t.Fatalf("Failed to get balance for userID %d: %v", userID, err)
	}

	expectedBalance := initialBalance + amount
	if balance != expectedBalance {
		t.Errorf("Unexpected balance: got %d, want %d", balance, expectedBalance)
	}
}

func TestCache_ConcurrentBalanceUpdates(t *testing.T) {
	cache := setupRedis()
	defer teardownRedis(cache)

	userID := 1
	initialBalance := 1000
	amount := 200

	// Ensure initial balance is set before retrieval
	success := cache.SetBalance(userID, initialBalance, false, false)
	if !success {
		t.Fatalf("Failed to set initial balance for userID %d", userID)
	}

	var wg sync.WaitGroup
	concurrentUpdates := 10

	for i := 0; i < concurrentUpdates; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.SetBalance(userID, amount+id, false, false)
		}(i)
	}

	wg.Wait()

	// Verify updated balance
	balance, err := cache.GetBalance(userID)
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	expectedBalance := initialBalance + (amount * concurrentUpdates)
	if balance != expectedBalance {
		t.Errorf("Unexpected balance: got %d, want %d", balance, expectedBalance)
	}
}

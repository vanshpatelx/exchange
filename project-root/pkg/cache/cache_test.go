package cache

import (
	"testing"
	"project/pkg/models"
	// "sync"
	"strings"
	"fmt"
	"log"
	"project/pkg/config"
)

func setupTestCache() (cache1, cache2 *Cache) {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cache1 = NewCache(config.REDIS_URL1)
	cache2 = NewCache(config.REDIS_URL2)

	cache1.initRedisClient()
	cache2.initRedisClient()
	return cache1, cache2
}

func TestNewStockAdd(t *testing.T) {
	_, cache2 := setupTestCache()

	// Initialize data
	stocks := []models.Stock{
		{TickerID: 1, Quantity: 100, LQ: 50, Price: 200},
		{TickerID: 2, Quantity: 200, LQ: 100, Price: 150},
		{TickerID: 3, Quantity: 100, LQ: 50, Price: 200},
	}

	// Build the string for stocks data
	var stockBuilder strings.Builder
	for _, stock := range stocks {
		// Format each stock entry as "TickerID,Quantity,LQ,Price"
		stockStr := fmt.Sprintf("%d,%d,%d,%d", stock.TickerID, stock.Quantity, stock.LQ, stock.Price)
		stockBuilder.WriteString(stockStr + ".")
	}

	// Remove the last period (not a comma)
	result := stockBuilder.String()
	if len(result) > 0 {
		result = result[:len(result)-1] // Removing the last period instead of comma
	}

	fmt.Print(result)

	// Set initial balances
	currentBalance := 10000
	lockedBalance := 5000
	balanceStr := fmt.Sprintf("%d,%d", currentBalance, lockedBalance)

	userId := 1
	cache1.InitializeData(userId, result)

	// Modify stock data
	stockModify := []models.Stock{
		{TickerID: 1, Quantity: 50, Price: 100},
		{TickerID: 2, Quantity: 100, Price: 100},
		{TickerID: 3, Quantity: 20, Price: 50},
	}

	// Set modified stocks data in cache
	for _, stock := range stockModify {
		success := cache.SetStock(userId, stock.TickerID, stock.Quantity, stock.Price, false, false)
		if !success {
			t.Errorf("Failed to set stock data for TickerID %d", stock.TickerID)
		}
	}

	// Retrieve and validate stocks
	retrievedStocks, err := cache.GetStocks(userId)
	if err != nil {
		t.Errorf("Failed to retrieve stocks: %v", err)
	}

	if len(retrievedStocks) != len(stockModify) { // Correctly compare the modified stocks
		t.Errorf("Expected %d stocks, got %d", len(stockModify), len(retrievedStocks))
	}

	for i, stock := range stockModify {
		if stock.TickerID != retrievedStocks[i].TickerID ||
			stock.Quantity != retrievedStocks[i].Quantity ||
			stock.LQ != retrievedStocks[i].LQ ||
			stock.Price != retrievedStocks[i].Price {
			t.Errorf("Stock mismatch at index %d: expected %+v, got %+v", i, stock, retrievedStocks[i])
		}
	}

	// Set balance data for user 1
	success := cache.SetBalance(userId+1, 1000, false, true)
	if !success {
		t.Errorf("Failed to set balance for user %d", userId+1)
	}

	// Retrieve and validate balance
	balance, locked, err := cache.GetBalance(userId+1)
	if err != nil {
		t.Errorf("Failed to retrieve balance: %v", err)
	}

	// Corrected balance check
	if balance != currentBalance+1000 {
		t.Errorf("Expected balance %d, got %d", currentBalance+1000, balance)
	}

	if locked != lockedBalance {
		t.Errorf("Expected locked balance %d, got %d", lockedBalance, locked)
	}
}

// func TestConcurrency(t *testing.T) {
// 	cache := setupTestCache()

// 	// Initialize stocks for user 1
// 	userId := 1
// 	stocks := []models.Stock{
// 		{TickerID: 101, Quantity: 100, LQ: 50, Price: 200},
// 		{TickerID: 102, Quantity: 200, LQ: 100, Price: 150},
// 	}

// 	// Set stocks data in cache
// 	for _, stock := range stocks {
// 		success := cache.SetStock(userId, stock.TickerID, stock.Quantity, stock.Price, false, false)
// 		if !success {
// 			t.Errorf("Failed to set stock data for TickerID %d", stock.TickerID)
// 		}
// 	}

// 	// Set initial balance data for user 1
// 	cache.SetBalance(userId, 1000, false, false)

// 	var wg sync.WaitGroup

// 	// Perform concurrent updates on stocks and balance
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()

// 			// Simulate a stock update
// 			tickerID := 101
// 			quantity := 10
// 			price := 250
// 			success := cache.SetStock(userId, tickerID, quantity, price, false, false)
// 			if !success {
// 				t.Errorf("Failed to update stock data for TickerID %d", tickerID)
// 			}

// 			// Simulate a balance update
// 			amount := 50
// 			success = cache.SetBalance(userId, amount, false, false)
// 			if !success {
// 				t.Errorf("Failed to update balance for user %d", userId)
// 			}
// 		}(i)
// 	}

// 	// Wait for all goroutines to finish
// 	wg.Wait()

// 	// Validate final stock and balance data
// 	retrievedStocks, err := cache.GetStocks(userId)
// 	if err != nil {
// 		t.Errorf("Failed to retrieve stocks: %v", err)
// 	}

// 	if len(retrievedStocks) != len(stocks) {
// 		t.Errorf("Expected %d stocks, got %d", len(stocks), len(retrievedStocks))
// 	}

// 	balance, err := cache.GetBalance(userId)
// 	if err != nil {
// 		t.Errorf("Failed to retrieve balance: %v", err)
// 	}

// 	if balance <= 0 {
// 		t.Errorf("Balance should be greater than 0, got %d", balance)
// 	}
// }

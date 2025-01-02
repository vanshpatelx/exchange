package cache_test

import (
	"context"
	"fmt"
	"project/pkg/cache"
	"project/pkg/models"
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
)

const (
	redisURL  = "localhost:6379" // Replace with your Redis URL if different
	redisURL2 = "localhost:6380"
)

func setupCache() (c1, c2 *cache.Cache) {
	c1 = cache.NewCache(redisURL)
	c2 = cache.NewCache(redisURL2)
	c2.InitializeStock(1, "101,50,10,10000")
	c1.InitializeBalance(1, "10000000,10000")
	return c1, c2
}

func TestCacheOperations(t *testing.T) {
	c1, c2 := setupCache()
	defer func() {
		c1.GetRedisClient().FlushAll(context.Background()) // Clean up after tests
		c2.GetRedisClient().FlushAll(context.Background()) // Clean up after tests
	}()

	t.Run("Test GetStocks", func(t *testing.T) {
		stocks, err := c2.GetStocks(1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(stocks))

		expectedStock := models.Stock{TickerID: 101, Quantity: 50, LQ: 10, Price: 10000}
		assert.Equal(t, expectedStock, stocks[0])
	})

	t.Run("Test SetStock Add Quantity", func(t *testing.T) {
		success := c2.SetStock(1, 101, 10, 10500, false, false)
		assert.True(t, success)

		stocks, err := c2.GetStocks(1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(stocks))

		expectedStock := models.Stock{TickerID: 101, Quantity: 60, LQ: 10, Price: 10083}
		assert.Equal(t, expectedStock, stocks[0])
	})

	t.Run("Test SetBalance Add Amount", func(t *testing.T) {
		amountToCheck := 20000

		currentAmountInitially, lockedAmountInitially, err := c1.GetBalance(1)
		assert.NoError(t, err)
		success := c1.SetBalance(1, amountToCheck, "SellOrderAddMoney")
		assert.True(t, success)

		currentAmount, lockedAmount, err := c1.GetBalance(1)
		assert.NoError(t, err)
		assert.Equal(t, currentAmountInitially+amountToCheck, currentAmount)
		assert.Equal(t, lockedAmountInitially, lockedAmount)
	})

	// t.Run("Test SetStock Rollback", func(t *testing.T) {
	// 	success := c2.SetStock(1, 101, -10, 95, true, false)
	// 	assert.True(t, success)

	// 	stocks, err := c2.GetStocks(1)
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, 1, len(stocks))

	// 	expectedStock := models.Stock{TickerID: 101, Quantity: 50, LQ: 20, Price: 100}
	// 	assert.Equal(t, expectedStock, stocks[0])
	// })

	// t.Run("Test SetBalance Rollback", func(t *testing.T) {
	// 	success := c1.SetBalance(1, -100)
	// 	assert.True(t, success)

	// 	currentAmount, lockedAmount, err := c1.GetBalance(1)
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, 1200, currentAmount) // Adjusted expected value
	// 	assert.Equal(t, 0, lockedAmount)
	// })

}

func BenchmarkCacheOperations(b *testing.B) {
	c1, c2 := setupCache()

	for i := 0; i < b.N; i++ {
		c2.SetStock(1, 101, 10, 105, false, false)
		c1.SetBalance(1, 200, "SellOrderAddMoney")
	}
}

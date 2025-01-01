package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"project/pkg/models"
	"strconv"
	"strings"
	"time"
)

type Cache struct {
	redisClient *redis.Client
	redisURL    string
}

func NewCache(redisURL string) *Cache {
	return &Cache{
		redisURL: redisURL,
	}
}

func (c *Cache) initRedisClient() {
	c.redisClient = redis.NewClient(&redis.Options{
		Addr: c.redisURL,
	})

	ctx := context.Background()

	_, err := c.redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	} else {
		fmt.Printf("Redis Connected to %s\n", c.redisURL)
	}
}

func (c *Cache) GetRedisClient() *redis.Client {
	if c.redisClient == nil {
		c.initRedisClient()
	}
	return c.redisClient
}

func (c *Cache) GetStocks(userId int) ([]models.Stock, error) {
	ctx := context.Background()
	client := c.GetRedisClient()

	stocksStr, err := client.Get(ctx, strconv.Itoa(userId)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	// Split the string by '.' to individual stock records
	stockEntries := strings.Split(stocksStr, ".")

	var stocks []models.Stock
	for _, entry := range stockEntries {
		// Split each stock record by ',' to extract fields
		fields := strings.Split(entry, ",")
		if len(fields) != 4 {
			return nil, fmt.Errorf("invalid stock entry: %s", entry)
		}

		field0, err := strconv.Atoi(fields[0]) // TickerID
		if err != nil {
			return nil, fmt.Errorf("invalid field0 in stock entry: %s", entry)
		}
		field1, err := strconv.Atoi(fields[1]) // Quantity
		if err != nil {
			return nil, fmt.Errorf("invalid field1 in stock entry: %s", entry)
		}
		field2, err := strconv.Atoi(fields[2]) // Locked Quantity (LQ)
		if err != nil {
			return nil, fmt.Errorf("invalid field2 in stock entry: %s", entry)
		}
		field3, err := strconv.Atoi(fields[3]) // Price
		if err != nil {
			return nil, fmt.Errorf("invalid field3 in stock entry: %s", entry)
		}

		stock := models.Stock{TickerID: field0, Quantity: field1, LQ: field2, Price: field3}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}
func (c *Cache) SetStock(userId int, tickerID int, quantity int, price int, rollback bool, settlement bool) bool {
	client := c.GetRedisClient()
	ctx := context.Background()

	// retry mechanism parameters
	retryDeadline := time.Now().Add(5 * time.Minute)
	var locked bool
	var err error

	// Retry acquiring the lock until 1 minute has passed
	for time.Now().Before(retryDeadline) {
		locked, err = c.LockResource(userId, strconv.Itoa(tickerID), 5*time.Minute)
		if err != nil {
			fmt.Printf("Error acquiring lock: %v\n", err)
			return false
		}
		if locked {
			break
		}

		// Wait before retrying
		fmt.Println("Resource is locked, retrying...")
		time.Sleep(2 * time.Second) // Retry after 2 seconds
	}

	// If lock was not acquired within 1 minute, return error
	if !locked {
		fmt.Println("Failed to acquire lock within 1 minute.")
		return false
	}

	defer func() {
		// lock is always released
		if unlockErr := c.UnlockResource(userId, strconv.Itoa(tickerID)); unlockErr != nil {
			fmt.Printf("Error releasing lock: %v\n", unlockErr)
		}
	}()

	// Retrieve current stock data
	stocksStr, err := client.Get(ctx, strconv.Itoa(userId)).Result()
	if err != nil {
		if err == redis.Nil {
			fmt.Printf("User's stock data not found: %d\n", userId)
			return false
		}
		fmt.Printf("Error fetching user's stock data: %v\n", err)
		return false
	}

	// Split the string by '.' to individual stock records
	stockEntries := strings.Split(stocksStr, ".")

	var updatedStocks []models.Stock
	var stockUpdated bool

	// Iterate to each stock
	for _, entry := range stockEntries {
		// Split each stock record by ',' to extract fields
		fields := strings.Split(entry, ",")
		if len(fields) != 4 {
			fmt.Printf("Invalid stock entry: %s\n", entry)
			return false
		}

		field0, err := strconv.Atoi(fields[0]) // TickerID
		if err != nil {
			fmt.Printf("Invalid field0 in stock entry: %s\n", entry)
			return false
		}

		field1, err := strconv.Atoi(fields[1]) // Quantity
		if err != nil {
			fmt.Printf("Invalid field1 in stock entry: %s\n", entry)
			return false
		}

		field2, err := strconv.Atoi(fields[2]) // Locked Quantity (LQ)
		if err != nil {
			fmt.Printf("Invalid field2 in stock entry: %s\n", entry)
			return false
		}

		field3, err := strconv.Atoi(fields[3]) // Price
		if err != nil {
			fmt.Printf("Invalid field3 in stock entry: %s\n", entry)
			return false
		}

		// Check is the same stock we're updating
		if field0 == tickerID {
			var updatedStock models.Stock
			if quantity < 0 && !rollback {
				// Real Sell Order: Change in locked quantity, price remains same
				updatedStock = models.Stock{
					TickerID: field0,
					Quantity: field1,
					LQ:       field2 - quantity,
					Price:    field3,
				}
			} else if quantity < 0 && rollback {
				// Rollback Sell Order: Price Change, deduct Quantity
				totalQuantity := field1
				totalPrice := field3
				newQuantity := totalQuantity - quantity
				newPrice := (totalPrice*totalQuantity - quantity*price) / newQuantity

				updatedStock = models.Stock{
					TickerID: field0,
					Quantity: newQuantity,
					LQ:       field2,
					Price:    newPrice,
				}
			} else if quantity > 0 && !rollback && settlement {
				// settlement only add left quantity to main file
				updatedStock = models.Stock{
					TickerID: field0,
					Quantity: field1 + quantity,
					LQ:       field2 - quantity,
					Price:    field3,
				}
			} else if quantity > 0 && !rollback {
				// Buy Order: Change in price and main Quantity
				newQuantity := field1 + quantity
				newPrice := (field1*field3 + quantity*price) / newQuantity

				updatedStock = models.Stock{
					TickerID: field0,
					Quantity: newQuantity,
					LQ:       field2,
					Price:    newPrice,
				}
			}

			updatedStocks = append(updatedStocks, updatedStock)
			stockUpdated = true
		} else {
			// Keep the current stock unchanged if it's not the one being updated
			updatedStocks = append(updatedStocks, models.Stock{
				TickerID: field0,
				Quantity: field1,
				LQ:       field2,
				Price:    field3,
			})
		}

	}

	if !stockUpdated {
		fmt.Printf("Stock not found: %d\n", tickerID)
		return false
	}

	// Prepare updated stock string
	var updatedStockEntries []string
	for _, stock := range updatedStocks {
		stockEntry := fmt.Sprintf("%d,%d,%d,%d", stock.TickerID, stock.Quantity, stock.LQ, stock.Price)
		updatedStockEntries = append(updatedStockEntries, stockEntry)
	}

	// Store the updated stock data back to Redis
	updatedStockStr := strings.Join(updatedStockEntries, ".")
	_, err = client.Set(ctx, strconv.Itoa(userId), updatedStockStr, 0).Result()
	if err != nil {
		fmt.Printf("Error saving updated stock data: %v\n", err)
		return false
	}

	return true
}

func (c *Cache) GetBalance(userId int) (int, int, error) {
	ctx := context.Background()
	client := c.GetRedisClient()

	balanceStr, err := client.Get(ctx, strconv.Itoa(userId)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	parts := strings.Split(balanceStr, ",")
	if len(parts) != 2 {
		fmt.Printf("Invalid balance format for user %d: %v\n", userId, balanceStr)
		return 0, 0, err
	}

	currentAmount, err := strconv.Atoi(parts[0]) // Amount
	if err != nil {
		fmt.Printf("Invalid amount format for user %d: %v\n", userId, err)
		return 0, 0, err
	}

	lockedAmount, err := strconv.Atoi(parts[1]) // Locked amount
	if err != nil {
		fmt.Printf("Invalid locked amount format for user %d: %v\n", userId, err)
		return 0, 0, err
	}

	return int(currentAmount), int(lockedAmount), nil
}

func (c *Cache) SetBalance(userId int, amount int, rollback bool, settlement bool) bool {
	client := c.GetRedisClient()
	ctx := context.Background()

	// retry mechanism parameters
	retryDeadline := time.Now().Add(5 * time.Minute)
	var locked bool
	var err error

	// Retry acquiring the lock until 5 minutes
	for time.Now().Before(retryDeadline) {
		locked, err = c.LockResource(userId, "balance", 5*time.Minute)
		if err != nil {
			fmt.Printf("Error acquiring lock: %v\n", err)
			return false
		}
		if locked {
			break
		}

		// Wait before retrying
		fmt.Println("Resource is locked, retrying...")
		time.Sleep(2 * time.Second) // Retry after 2 seconds
	}

	// If lock was not acquired within 5 minutes, return error
	if !locked {
		fmt.Println("Failed to acquire lock within 5 minutes.")
		return false
	}

	defer func() {
		// lock is always released
		if unlockErr := c.UnlockResource(userId, "balance"); unlockErr != nil {
			fmt.Printf("Error releasing lock: %v\n", unlockErr)
		}
	}()

	// Retrieve current balance and locked amount from Redis
	key := strconv.Itoa(userId)
	balanceStr, err := client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			fmt.Printf("Balance not found for user %d\n", userId)
		} else {
			fmt.Printf("Error fetching balance: %v\n", err)
		}
		return false
	}

	// Parse the current balance and locked amount
	parts := strings.Split(balanceStr, ",")
	if len(parts) != 2 {
		fmt.Printf("Invalid balance format for user %d: %v\n", userId, balanceStr)
		return false
	}

	currentAmount, err := strconv.Atoi(parts[0]) // Amount
	if err != nil {
		fmt.Printf("Invalid amount format for user %d: %v\n", userId, err)
		return false
	}

	lockedAmount, err := strconv.Atoi(parts[1]) // Locked amount
	if err != nil {
		fmt.Printf("Invalid locked amount format for user %d: %v\n", userId, err)
		return false
	}

	var newAmount int
	var newLockedAmount int

	if amount < 0 && !rollback {
		// sell Order: add amount
		newAmount = currentAmount + amount
	} else if amount < 0 && rollback {
		// buy Order Rollback: add to lockAmount
		newLockedAmount = lockedAmount + amount
	} else if amount > 0 && !rollback {
		// buy Order: less to lockAmount
		newLockedAmount = lockedAmount - amount
	} else if amount > 0 && rollback {
		// sell Order Rollback: deduct amount
		newAmount = currentAmount - amount
	} else if amount > 0 && !rollback && settlement {
		//  buyer's Settlement => lock remove and add money to main balance
		newAmount = currentAmount + amount
		newLockedAmount = lockedAmount - amount
	} else if amount > 0 && rollback && settlement {
		//  buyer's Settlement Rollback => lock remove and add money to main balance
		newAmount = currentAmount - amount
		newLockedAmount = lockedAmount + amount
	}

	updatedBalance := fmt.Sprintf("%d,%d", newAmount, newLockedAmount)
	_, err = client.Set(ctx, key, updatedBalance, 0).Result()
	if err != nil {
		fmt.Printf("Error updating balance for user %d: %v\n", userId, err)
		return false
	}

	return true
}

func (c *Cache) LockResource(userId int, resourceId string, lockTimeout time.Duration) (bool, error) {
	ctx := context.Background()
	client := c.GetRedisClient()

	lockKey := fmt.Sprintf("lock:%d:%s", userId, resourceId)

	success, err := client.SetNX(ctx, lockKey, "locked", lockTimeout).Result()
	if err != nil {
		return false, fmt.Errorf("error acquiring lock: %v", err)
	}

	return success, nil
}

func (c *Cache) UnlockResource(userId int, resourceId string) error {
	ctx := context.Background()
	client := c.GetRedisClient()

	lockKey := fmt.Sprintf("lock:%d:%s", userId, resourceId)

	_, err := client.Del(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("error releasing lock: %v", err)
	}

	return nil
}

func (c *Cache) InitializeStock(userId int, stocksStr string) {
	ctx := context.Background()
	client := c.GetRedisClient()

	client.Set(ctx, strconv.Itoa(userId), stocksStr, 0).Err()

	fmt.Printf("Initialized data for user %d\n", userId)
}

func (c *Cache) InitializeBalance(userId int, balanceStr string) {
	ctx := context.Background()
	client := c.GetRedisClient()

	client.Set(ctx, strconv.Itoa(userId), balanceStr, 0).Err()

	fmt.Printf("Initialized data for user %d\n", userId)
}


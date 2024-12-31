package cache

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"github.com/go-redis/redis/v8"
)

type Cache struct {
	redisClient *redis.Client
	redisURL    string
	once        sync.Once
}

func NewCache(redisURL string) *Cache {
	return &Cache{
		redisURL: redisURL,
	}
}

func (c *Cache) initRedisClient() {
	c.once.Do(func() {
		c.redisClient = redis.NewClient(&redis.Options{
			Addr:     c.redisURL,
		})

		ctx := context.Background()

		_, err := c.redisClient.Ping(ctx).Result()
		if err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}else{
			fmt.Print("Redis Connected")
		}
	})
}

func (c *Cache) GetRedisClient() *redis.Client {
	c.initRedisClient()
	return c.redisClient
}
func (c *Cache) GetBalance(userId int) (int, error) {
	ctx := context.Background()
    client := c.GetRedisClient()

    userKey := strconv.Itoa(userId)

    balanceStr, err := client.Get(ctx, userKey).Result()
    if err != nil {
        if err == redis.Nil {
            return 0, nil
        }
        return 0, err
    }

    balance, err := strconv.ParseInt(balanceStr, 10, 64) // Base 10, 64-bit
    if err != nil {
        return 0, fmt.Errorf("failed to parse balance for user %d: %w", userId, err)
    }

    return int(balance), nil
}

func (c *Cache) SetBalance(userId int, balance int) error {
	ctx := context.Background()
	client := c.GetRedisClient()

	userKey := strconv.Itoa(userId)
	userBalance := strconv.Itoa(balance)

	// Lua script =>  If the key does not exist, it creates it with the balance.
	luaScript := `
		-- KEYS[1] = userKey
		-- ARGV[1] = userBalance
		redis.call("SET", KEYS[1], ARGV[1])
		return 1
	`

	err := client.Eval(ctx, luaScript, []string{userKey}, userBalance).Err()
	if err != nil {
		return err
	}

	return nil
}

func settleBalance(redisClient *redis.Client, key string, amount int) bool {
	ctx := context.Background()

	luaScript := `
		local balance = redis.call("GET", KEYS[1])
		if not balance then
			return redis.error_reply("Balance not found")
		end
		balance = tonumber(balance)
		local newBalance = balance + tonumber(ARGV[1])
		if newBalance < 0 then
			return redis.error_reply("Balance cannot be less than zero")
		end
		redis.call("SET", KEYS[1], newBalance)
		return tostring(newBalance)
	`

	_, err := redisClient.Eval(ctx, luaScript, []string{key}, amount).Result()
	if err != nil {
		fmt.Printf("Error executing Lua script: %v\n", err)
		return false
	}

	return true
}

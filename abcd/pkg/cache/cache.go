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

func (c *Cache) GetBalance(ctx context.Context, userId string) (float64, error) {
	client := c.GetRedisClient()

	balanceStr, err := client.Get(ctx, userId).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (c *Cache) SetBalance(ctx context.Context, userId string, balance float64) error {
	client := c.GetRedisClient()

	balanceStr := strconv.FormatFloat(balance, 'f', -1, 64)

	err := client.Set(ctx, userId, balanceStr, 0).Err()
	if err != nil {
		return err
	}

	return nil
}


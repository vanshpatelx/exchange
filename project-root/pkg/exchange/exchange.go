package exchange

import (
	"project/pkg/cache"
	"project/pkg/models"
	"project/pkg/stockmanager"
	"project/pkg/pubsub"
	"sync"
)

type Exchange struct {
	StockManagers map[int]*stockmanager.StockManager
	mu            sync.RWMutex
	cache         *cache.Cache
	pubsub        *pubsub.PubSub
}

func NewExchange(redisURL string, pubsub *pubsub.PubSub, cache *cache.Cache) *Exchange {
	return &Exchange{
		StockManagers: make(map[int]*stockmanager.StockManager),
		cache:         cache,
		pubsub:        pubsub,
	}
}

func (e *Exchange) AddStock(Ticker int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.StockManagers[Ticker]; !exists {
		e.StockManagers[Ticker] = stockmanager.NewStockManager(Ticker, e.cache, e.pubsub)
	}
}

func (e *Exchange) PlaceOrder(order *models.Order) {
	e.mu.Lock()

	if manager, exists := e.StockManagers[order.Ticker]; exists {
		go manager.PlaceOrder(order)
	} else {
		e.AddStock(order.Ticker)
		manager := e.StockManagers[order.Ticker]
		go manager.PlaceOrder(order)
	}

	e.mu.Unlock()
}

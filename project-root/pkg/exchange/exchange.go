package exchange

import (
	"project/pkg/cache"
	"project/pkg/models"
	"project/pkg/pubsub"
	"project/pkg/stockmanager"
	"sync"
)

type Exchange struct {
	StockManagers map[int]*stockmanager.StockManager
	mu            sync.RWMutex
	cache1        *cache.Cache
	cache2        *cache.Cache
	pubsub        *pubsub.PubSub
	config        *models.Config
}

func NewExchange(pubsub *pubsub.PubSub, cache1 *cache.Cache, cache2 *cache.Cache, config *models.Config) *Exchange {
	return &Exchange{
		StockManagers: make(map[int]*stockmanager.StockManager),
		cache1:        cache1,
		cache2:        cache2,
		pubsub:        pubsub,
		config:        config,
	}
}

func (e *Exchange) AddStock(Ticker int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.StockManagers[Ticker]; !exists {
		e.StockManagers[Ticker] = stockmanager.NewStockManager(Ticker, e.cache1, e.cache2, e.pubsub, e.config)
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

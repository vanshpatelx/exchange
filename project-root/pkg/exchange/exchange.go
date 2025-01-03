package exchange

import (
	"project/pkg/cache"
	"project/pkg/models"
	"project/pkg/pubsub"
	"project/pkg/stockmanager"
	"sync"
	"log"
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
	if _, exists := e.StockManagers[Ticker]; !exists {
		e.StockManagers[Ticker] = stockmanager.NewStockManager(Ticker, e.cache1, e.cache2, e.pubsub, e.config)
	}
}

func (e *Exchange) PlaceOrder(order *models.Order) {
	e.mu.Lock()
	if manager, exists := e.StockManagers[order.Ticker]; exists {
		manager.PlaceOrder(order)
	} else {
		e.AddStock(order.Ticker)
		manager := e.StockManagers[order.Ticker]
		go manager.PlaceOrder(order)
		log.Printf("Added Stock:%d", order.Ticker)
	}
	log.Printf("PlaceOrder: Quantity:%d  Price:%d", order.Quantity, order.Price)
	e.mu.Unlock()
}

func (e *Exchange) Settlement(Ticker int) {
	e.mu.Lock()

	if manager, exists := e.StockManagers[Ticker]; exists {
		var wg sync.WaitGroup

		wg.Add(1)

		go func() {
			defer wg.Done() // Decrease the counter when the goroutine finishes
			manager.LeftSettlement(Ticker)
		}()

		wg.Wait()

		delete(e.StockManagers, Ticker)
	}

	e.mu.Unlock()
}

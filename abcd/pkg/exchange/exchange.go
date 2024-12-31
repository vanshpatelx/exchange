package exchange

import (
	"project/pkg/models"
	"project/pkg/stockmanager"
	"sync"
)

type Exchange struct {
	StockManagers map[string]*stockmanager.StockManager
	mu            sync.RWMutex
}

func NewExchange() *Exchange {
	return &Exchange{
		StockManagers: make(map[string]*stockmanager.StockManager),
	}
}

func (e *Exchange) AddStock(stockName string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.StockManagers[stockName]; !exists {
		e.StockManagers[stockName] = stockmanager.NewStockManager(stockName)
	}
}

func (e *Exchange) PlaceOrder(order *models.Order) {
	e.mu.Lock()

	if manager, exists := e.StockManagers[order.Stock]; exists {
		go manager.PlaceOrder(order)
	} else {
		e.AddStock(order.Stock)
		manager := e.StockManagers[order.Stock]
		go manager.PlaceOrder(order)
	}

	e.mu.Unlock()
}


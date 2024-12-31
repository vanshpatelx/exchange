package stockmanager

import (
	// "log"
	"project/pkg/models"
	"project/pkg/queue"
	"sync"
)

type StockManager struct {
	stockName string
	BuyQueue  *queue.PriorityQueue
	SellQueue *queue.PriorityQueue
	mu        sync.Mutex 
}

func NewStockManager(stockName string) *StockManager {
	return &StockManager{
		stockName: stockName,
		BuyQueue:  queue.NewPriorityQueue(),
		SellQueue: queue.NewPriorityQueue(),
	}
}

func (sm *StockManager) PlaceOrder(order *models.Order) {
	sm.mu.Lock() 
	defer sm.mu.Unlock()

	if order.Type == "buy" {
		sm.BuyQueue.Enqueue(order)
	} else if order.Type == "sell" {
		sm.SellQueue.Enqueue(order)
	}

	go sm.MatchOrders()
}

func (sm *StockManager) MatchOrders() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for {
		buyOrder := sm.BuyQueue.Peek()
		sellOrder := sm.SellQueue.Peek()

		if buyOrder != nil && sellOrder != nil && buyOrder.Price >= sellOrder.Price {
			sm.SettleTrade(buyOrder, sellOrder)
			sm.BuyQueue.Dequeue()
			sm.SellQueue.Dequeue()
		} else {
			break
		}
	}
}

func (sm *StockManager) SettleTrade(buyOrder, sellOrder *models.Order) {
	// log.Printf("Settled: %d of %s - Buy Price: %d, Sell Price: %d", buyOrder.Quantity, sm.stockName, buyOrder.Price, sellOrder.Price)

	go sm.cacheTradeData(buyOrder, sellOrder)
	go sm.publishTradeEvent(buyOrder, sellOrder)
}

func (sm *StockManager) cacheTradeData(buyOrder, sellOrder *models.Order) {
	// log.Printf("Caching trade data for %s: %d at %d", sm.stockName, buyOrder.Quantity, buyOrder.Price)
}

func (sm *StockManager) publishTradeEvent(buyOrder, sellOrder *models.Order) {
	// log.Printf("Publishing trade event for %s: %d at %d", sm.stockName, buyOrder.Quantity, buyOrder.Price)
}

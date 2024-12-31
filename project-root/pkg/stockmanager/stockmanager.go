package stockmanager

import (
	"encoding/json"
	"fmt"
	"project/pkg/cache"
	"project/pkg/idgen"
	"project/pkg/models"
	"project/pkg/pubsub"
	"project/pkg/queue"
	"sync"
)

type StockManager struct {
	Ticker    int
	BuyQueue  *queue.PriorityQueue
	SellQueue *queue.PriorityQueue
	mu        sync.Mutex
	cache1    *cache.Cache
	cache2    *cache.Cache
	pubsub    *pubsub.PubSub
	config    *models.Config
}

func NewStockManager(Ticker int, cache1 *cache.Cache, cache2 *cache.Cache, pubsub *pubsub.PubSub, config *models.Config) *StockManager {
	return &StockManager{
		Ticker:    Ticker,
		BuyQueue:  queue.NewPriorityQueue(),
		SellQueue: queue.NewPriorityQueue(),
		cache1:    cache1,
		cache2:    cache2,
		pubsub:    pubsub,
		config:    config,
	}
}

func (sm *StockManager) PlaceOrder(order *models.Order) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if order.Type == 1 {
		sm.BuyQueue.Enqueue(order)
	} else if order.Type == 0 {
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
			tradeQuantity := min(buyOrder.Quantity, sellOrder.Quantity)

			success := sm.SettlePartialTrade(buyOrder, sellOrder, tradeQuantity, sellOrder.Price)
			if success {
				buyOrder.Quantity -= tradeQuantity
				sellOrder.Quantity -= tradeQuantity

				if buyOrder.Quantity == 0 {
					sm.BuyQueue.Dequeue()
				}
				if sellOrder.Quantity == 0 {
					sm.SellQueue.Dequeue()
				}
			} else {
				break
			}
		} else {
			break
		}
	}
}

func (sm *StockManager) publishTradeEvent(buyOrder, sellOrder *models.Order, tradeQuantity int, tradePrice int) {
	id := idgen.GenerateUniqueId()

	trade := map[string]interface{}{
		"id":       id,
		"bOId":     buyOrder.Id,
		"sOId":     sellOrder.Id,
		"quantity": tradeQuantity,
		"price":    tradePrice,
		"bId":      buyOrder.User,
		"sId":      sellOrder.User,
		"ticker":   buyOrder.Ticker,
	}

	message, err := json.Marshal(trade)
	if err != nil {
		fmt.Printf("Error serializing trade event to JSON: %s\n", err)
		return
	}

	err = sm.pubsub.Publish(sm.config.PUBLISHER_EXCHANGE, sm.config.PUBLISHER_ROUTING_KEY, string(message))
	if err != nil {
		fmt.Printf("Error publishing trade event: %s\n", err)
	}
}
func (sm *StockManager) SettlePartialTrade(buyOrder, sellOrder *models.Order, tradeQuantity int, tradePrice int) bool {
	amount := tradePrice * tradeQuantity

	// buyer's balance
	if !sm.cache1.SetBalance(buyOrder.User, amount, false) {
		return false
	}

	// seller's balance
	if !sm.cache1.SetBalance(sellOrder.User, -amount, false) {
		sm.cache1.SetBalance(buyOrder.User, -amount, true) // Rollback
		return false
	}

	// buyer's holdings
	if !sm.cache2.SetStock(buyOrder.User, buyOrder.Id, tradeQuantity, tradePrice, false) {
		sm.cache1.SetBalance(buyOrder.User, -amount, true) // Rollback
		sm.cache1.SetBalance(sellOrder.User, amount, true)
		return false
	}

	// seller's holdings
	if !sm.cache2.SetStock(sellOrder.User, sellOrder.Id, -tradeQuantity, tradePrice, false) {
		sm.cache1.SetBalance(buyOrder.User, -amount, true) // Rollback
		sm.cache1.SetBalance(sellOrder.User, amount, true)
		sm.cache2.SetStock(buyOrder.User, buyOrder.Id, -tradeQuantity, tradePrice, true)
		return false
	}

	sm.publishTradeEvent(buyOrder, sellOrder, tradeQuantity, tradePrice)

	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package stockmanager

import (
	"encoding/json"
	"fmt"
	"project/pkg/cache"
	"project/pkg/models"
	"project/pkg/pubsub"
	"project/pkg/queue"
	"sync"
    "project/pkg/idgen"
)

type StockManager struct {
	Ticker    int
	BuyQueue  *queue.PriorityQueue
	SellQueue *queue.PriorityQueue
	mu        sync.Mutex
	cache     *cache.Cache
	pubsub    *pubsub.PubSub
}

func NewStockManager(Ticker int, cache *cache.Cache, pubsub *pubsub.PubSub) *StockManager {
	return &StockManager{
		Ticker:    Ticker,
		BuyQueue:  queue.NewPriorityQueue(),
		SellQueue: queue.NewPriorityQueue(),
		cache:     cache,
		pubsub:    pubsub,
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

	err = sm.pubsub.Publish("trade_exchange", "trade_event", string(message))
	if err != nil {
		fmt.Printf("Error publishing trade event: %s\n", err)
	}
}

func (sm *StockManager) SettlePartialTrade(buyOrder, sellOrder *models.Order, tradeQuantity int, tradePrice int) bool {
	buyAmount := tradePrice * tradeQuantity
	sellAmount := tradePrice * tradeQuantity


	

	buyBalance, err := sm.cache.GetBalance(buyOrder.User)
	if err != nil {
		fmt.Printf("Error fetching buyer balance: %v\n", err)
		return false
	}

	sellBalance, err := sm.cache.GetBalance(sellOrder.User)
	if err != nil {
		fmt.Printf("Error fetching seller balance: %v\n", err)
		return false
	}

	if buyBalance < int(buyAmount) {
		fmt.Println("Insufficient balance for trade")
		return false
	}

	if err := sm.cache.SetBalance(buyOrder.User, buyBalance-int(buyAmount)); err != nil {
		fmt.Printf("Error updating buyer balance: %v\n", err)
		return false
	}

	if err := sm.cache.SetBalance(sellOrder.User, sellBalance+int(sellAmount)); err != nil {
		fmt.Printf("Error updating seller balance: %v\n", err)
		sm.cache.SetBalance(buyOrder.User, buyBalance) // Attempt rollback
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

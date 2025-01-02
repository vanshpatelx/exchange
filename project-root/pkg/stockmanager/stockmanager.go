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
	"log"
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
		log.Printf("BuyQueue: Quantity:%d  Price:%d", order.Quantity, order.Price)
	} else if order.Type == 0 {
		sm.SellQueue.Enqueue(order)
		log.Printf("SellQueue: Quantity:%d  Price:%d", order.Quantity, order.Price)
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

			log.Printf("OrderMatched: Quantity:%d  Price:%d", tradeQuantity, sellOrder.Price)

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
	if !sm.cache1.SetBalance(buyOrder.User, amount, "SellOrderDeductLockAmount") {
		return false
	}

	// seller's balance
	if !sm.cache1.SetBalance(sellOrder.User, amount, "SellOrderAddMoney") {
		sm.cache1.SetBalance(buyOrder.User, amount, "BuyOrderRollback") // Rollback
		return false
	}

	// buyer's holdings
	if !sm.cache2.SetStock(buyOrder.User, buyOrder.Id, tradeQuantity, tradePrice, false, false) {
		sm.cache1.SetBalance(buyOrder.User, amount, "BuyOrderRollback") // Rollback
		sm.cache1.SetBalance(sellOrder.User, amount, "SellOrderRollback")
		return false
	}

	// seller's holdings
	if !sm.cache2.SetStock(sellOrder.User, sellOrder.Id, -tradeQuantity, tradePrice, false, false) {
		sm.cache1.SetBalance(buyOrder.User, amount, "BuyOrderRollback") // Rollback
		sm.cache1.SetBalance(sellOrder.User, amount, "SellOrderRollback")
		sm.cache2.SetStock(buyOrder.User, buyOrder.Id, -tradeQuantity, tradePrice, true, false)
		return false
	}

	log.Printf("Settlement:%d %d %d", sellOrder.Id, buyOrder.Id, sellOrder.Price)
	sm.publishTradeEvent(buyOrder, sellOrder, tradeQuantity, tradePrice)

	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (sm *StockManager) LeftSettlement(Ticker int) {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    // Clean buy Queue
    for {
        buyOrder := sm.BuyQueue.Peek()
        if buyOrder == nil {
            break
        }

        amount := buyOrder.Price * buyOrder.Quantity
        if !sm.cache1.SetBalance(buyOrder.User, amount, false, true) {
            fmt.Printf("Failed to update balance for user %d\n", buyOrder.User)
        }
		sm.BuyQueue.Pop() // Remove processed order
    }

    // Clean sell Queue
    for {
        sellOrder := sm.SellQueue.Peek()
        if sellOrder == nil {
            break
        }

        if !sm.cache2.SetStock(sellOrder.User, Ticker, sellOrder.Quantity, sellOrder.Price, false, true) {
            fmt.Printf("Failed to update stock for user %d\n", sellOrder.User)
        }
		sm.SellQueue.Pop() // Remove processed order
    }
}


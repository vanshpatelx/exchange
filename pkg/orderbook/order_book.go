package orderbook

import (
	"container/heap"
	// "fmt"
	"sync"
)

type OrderBook struct {
	buyLevels   *PriorityQueue
	sellLevels  *PriorityQueue
	priceLevels map[uint64]*PriceLevel // Stores PriceLevels by Price
	mu          sync.Mutex
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		buyLevels:   NewPriorityQueue(true),
		sellLevels:  NewPriorityQueue(false),
		priceLevels: make(map[uint64]*PriceLevel),
	}
}

func (ob *OrderBook) AddOrder(order *Order) {
	ob.mu.Lock() // Lock the OrderBook for safe concurrent access
	defer ob.mu.Unlock()

	if !order.Type { // MARKET
		// fmt.Print("\nMARKET", order)
		ob.matchMarketOrder(order)
	} else { // LIMIT
		// fmt.Print("\nLIMIT", order)
		ob.addLimitOrder(order)
	}
}

func (ob *OrderBook) addLimitOrder(order *Order) {
	var pq *PriorityQueue
	var oppositeSide []*PriorityQueue

	if order.Side {
		pq = ob.buyLevels
		oppositeSide = []*PriorityQueue{ob.sellLevels}
	} else {
		pq = ob.sellLevels
		oppositeSide = []*PriorityQueue{ob.buyLevels}
	}

	// Check if there are existing matching orders at the same price level
	priceLevel, exists := ob.priceLevels[uint64(order.Price)]
	if !exists {
		// No matching orders, add this limit order to the book
		priceLevel = NewPriceLevel(order.Price)
		ob.priceLevels[uint64(order.Price)] = priceLevel
		heap.Push(pq, priceLevel)
	}

	// Try to match this order with available orders in the opposite side (buy/sell)
	if len(oppositeSide) > 0 {
		ob.matchLimitOrder(order, oppositeSide)
	} else {
		// Add the order to the price level if no match was found
		priceLevel.AddOrder(order)
	}
}

func (ob *OrderBook) addLimitOrderToBook(order *Order) {
	// Add the limit order to the order book if it wasn't matched
	var pq *PriorityQueue
	if order.Side { // BUY
		pq = ob.buyLevels
	} else { // SELL
		pq = ob.sellLevels
	}

	priceLevel, exists := ob.priceLevels[uint64(order.Price)]
	if !exists {
		// No matching price level, create a new one
		priceLevel = NewPriceLevel(order.Price)
		ob.priceLevels[uint64(order.Price)] = priceLevel
		heap.Push(pq, priceLevel)
	}

	// Add the order to the price level
	priceLevel.AddOrder(order)
}

func (ob *OrderBook) matchLimitOrder(order *Order, oppositeSide []*PriorityQueue) {
	// Match the limit order against the opposite side's price levels
	for _, pq := range oppositeSide {
		// Loop through the opposite side price levels
		for pq.Len() > 0 && order.Quantity > 0 {
			bestLevel := pq.levels[0]

			// If the price level matches (sell for buy, buy for sell)
			if (order.Side && bestLevel.Price <= order.Price) || (!order.Side && bestLevel.Price >= order.Price) {
				// Match orders from this price level
				for _, existingOrder := range bestLevel.Orders {
					if order.Quantity <= 0 {
						break
					}

					// order fully matched
					if existingOrder.Quantity <= order.Quantity {
						order.Quantity -= existingOrder.Quantity
						bestLevel.RemoveOrder(existingOrder.ID)
						// fmt.Printf("Trade complete: Order ID %d and Order ID %d, Quantity: %d, Price: %d\n",
						// 	existingOrder.ID, order.ID, existingOrder.Quantity, bestLevel.Price)
					} else {
						// Partial match
						existingOrder.Quantity -= order.Quantity
						// fmt.Printf("Trade complete: Order ID %d and Order ID %d, Quantity: %d, Price: %d\n",
							// existingOrder.ID, order.ID, order.Quantity, bestLevel.Price)
						order.Quantity = 0
					}
				}

				// If no orders remain at this price level, remove the price level
				if len(bestLevel.Orders) == 0 {
					heap.Pop(pq)
					delete(ob.priceLevels, uint64(bestLevel.Price))
				}
			} else {
				// No match at this price level
				break
			}
		}
	}

	// If the order still has quantity remaining, add it to the book
	if order.Quantity > 0 {
		ob.addLimitOrderToBook(order)
	}
}

func (ob *OrderBook) matchMarketOrder(order *Order) {
	var pq *PriorityQueue

	if order.Side { // BUY market order
		pq = ob.sellLevels
	} else { // SELL market order
		pq = ob.buyLevels
	}

	// Match the market order with available orders
	for pq.Len() > 0 && order.Quantity > 0 {
		bestLevel := pq.levels[0]
		temp := order.Quantity

		for _, existingOrder := range bestLevel.Orders {
			if order.Quantity <= 0 {
				break
			}

			// Execute the match and reduce quantities accordingly
			if existingOrder.Quantity <= order.Quantity {
				order.Quantity -= existingOrder.Quantity
				bestLevel.RemoveOrder(existingOrder.ID)
				// fmt.Printf("Trade complete: Order ID %d and Order ID %d, Quantity: %d, Price: %d\n",
				// 	existingOrder.ID, order.ID, existingOrder.Quantity, bestLevel.Price)
			} else {
				existingOrder.Quantity -= order.Quantity
				// fmt.Printf("Trade complete: Order ID %d and Order ID %d, Quantity: %d, Price: %d\n",
				// 	existingOrder.ID, order.ID, order.Quantity, bestLevel.Price)
				order.Quantity = 0

			}
		}

		bestLevel.Volume -= uint64(temp)

		// Remove empty price levels
		if len(bestLevel.Orders) == 0 {
			heap.Pop(pq)
			delete(ob.priceLevels, uint64(bestLevel.Price))
		}
	}
}

func (ob *OrderBook) BuyLevels() *PriorityQueue {
	return ob.buyLevels
}

func (ob *OrderBook) SellLevels() *PriorityQueue {
	return ob.sellLevels
}

package orderbook

import (
	"testing"
)

func TestNewOrderBook(t *testing.T) {
	orderBook := NewOrderBook()
	if orderBook == nil {
		t.Fatal("NewOrderBook() should not return nil")
	}
	if len(orderBook.buyLevels.levels) != 0 {
		t.Fatal("Newly created buy levels should be empty")
	}
	if len(orderBook.sellLevels.levels) != 0 {
		t.Fatal("Newly created sell levels should be empty")
	}
}

func TestAddOrder_LimitOrderBuy(t *testing.T) {
	orderBook := NewOrderBook()
	order := &Order{
		ID:       1,
		Type:     true, // LIMIT order
		Side:     true, // BUY
		Price:    100,
		Quantity: 10,
	}
	orderBook.AddOrder(order)

	if len(orderBook.buyLevels.levels) != 1 {
		t.Fatalf("Expected 1 buy level, got %d", len(orderBook.buyLevels.levels))
	}
	if orderBook.buyLevels.levels[0].Price != 100 {
		t.Fatalf("Expected price level 100, got %d", orderBook.buyLevels.levels[0].Price)
	}
	if orderBook.buyLevels.levels[0].Volume != 10 {
		t.Fatalf("Expected volume 10, got %d", orderBook.buyLevels.levels[0].Volume)
	}
}

func TestAddOrder_LimitOrderSell(t *testing.T) {
	orderBook := NewOrderBook()
	order := &Order{
		ID:       1,
		Type:     true,  // LIMIT order
		Side:     false, // SELL
		Price:    150,
		Quantity: 5,
	}
	orderBook.AddOrder(order)

	if len(orderBook.sellLevels.levels) != 1 {
		t.Fatalf("Expected 1 sell level, got %d", len(orderBook.sellLevels.levels))
	}
	if orderBook.sellLevels.levels[0].Price != 150 {
		t.Fatalf("Expected price level 150, got %d", orderBook.sellLevels.levels[0].Price)
	}
	if orderBook.sellLevels.levels[0].Volume != 5 {
		t.Fatalf("Expected volume 5, got %d", orderBook.sellLevels.levels[0].Volume)
	}
}

func TestAddOrder_MarketOrderBuy(t *testing.T) {
	orderBook := NewOrderBook()
	// Add a sell limit order
	sellOrder := &Order{
		ID:       1,
		Type:     true,  // LIMIT
		Side:     false, // SELL
		Price:    100,
		Quantity: 10,
	}
	orderBook.AddOrder(sellOrder)

	// Add a market buy order
	marketOrder := &Order{
		ID:       2,
		Type:     false, // MARKET
		Side:     true,  // BUY
		Price:    0,     // Ignored for MARKET orders
		Quantity: 5,
	}
	orderBook.AddOrder(marketOrder)

	if orderBook.sellLevels.levels[0].Volume != 5 {
		t.Fatalf("Expected sell level volume to decrease to 5, got %d", orderBook.sellLevels.levels[0].Volume)
	}
}

func TestAddOrder_MarketOrderSell(t *testing.T) {
	orderBook := NewOrderBook()
	// Add a buy limit order
	buyOrder := &Order{
		ID:       1,
		Type:     true, // LIMIT
		Side:     true, // BUY
		Price:    100,
		Quantity: 10,
	}
	orderBook.AddOrder(buyOrder)

	// Add a market sell order
	marketOrder := &Order{
		ID:       2,
		Type:     false, // MARKET
		Side:     false, // SELL
		Price:    0,     // Ignored for MARKET orders
		Quantity: 5,
	}
	orderBook.AddOrder(marketOrder)

	if orderBook.buyLevels.levels[0].Volume != 5 {
		t.Fatalf("Expected buy level volume to decrease to 5, got %d", orderBook.buyLevels.levels[0].Volume)
	}
}

func TestOrderMatching(t *testing.T) {
	orderBook := NewOrderBook()
	// Add a buy and sell limit order that should match
	buyOrder := &Order{
		ID:       1,
		Type:     true, // LIMIT
		Side:     true, // BUY
		Price:    100,
		Quantity: 10,
	}
	orderBook.AddOrder(buyOrder)

	sellOrder := &Order{
		ID:       2,
		Type:     true,  // LIMIT
		Side:     false, // SELL
		Price:    100,
		Quantity: 10,
	}
	orderBook.AddOrder(sellOrder)

	if len(orderBook.buyLevels.levels) != 0 {
		t.Fatal("Expected buy levels to be empty after matching")
	}
	if len(orderBook.sellLevels.levels) != 0 {
		t.Fatal("Expected sell levels to be empty after matching")
	}
}

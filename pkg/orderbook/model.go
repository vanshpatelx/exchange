package orderbook

type Order struct {
	ID       uint64
	Type     bool  // "LIMIT => 1" or "MARKET => 0"
	Side     bool  // "BUY => 1" or "SELL => 0"
	Price    uint32 // Price for LIMIT orders
	Quantity uint32 // Quantity to buy or sell
}
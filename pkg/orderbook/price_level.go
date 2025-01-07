package orderbook

type PriceLevel struct {
	Price    uint32
	Volume   uint64
	Orders   map[uint64]*Order
	OrderIDs []uint64 // Maintains insertion order
}

// AddOrder
func (pl *PriceLevel) AddOrder(order *Order) {
	if _, exists := pl.Orders[order.ID]; !exists {
		pl.OrderIDs = append(pl.OrderIDs, order.ID)
		pl.Orders[order.ID] = order
		pl.Volume += uint64(order.Quantity)
	}
}

// RemoveOrder
func (pl *PriceLevel) RemoveOrder(orderId uint64) {
	if order, exists := pl.Orders[orderId]; exists {
		pl.Volume -= uint64(order.Quantity)
		delete(pl.Orders, orderId)
		for i, id := range pl.OrderIDs {
			if id == orderId {
				pl.OrderIDs = append(pl.OrderIDs[:i], pl.OrderIDs[i+1:]...)
				break
			}
		}
	}
}

// GetOrders - We use OrderIDs because we want FIFO order
func (pl *PriceLevel) GetOrders() []*Order {
	var ordersSlice []*Order

	for _, order := range pl.OrderIDs {
		ordersSlice = append(ordersSlice, pl.Orders[order])
	}
	return ordersSlice
}

func NewPriceLevel(price uint32) *PriceLevel {
	return &PriceLevel{
		Price:    price,
		Volume:   0,
		Orders:   make(map[uint64]*Order),
		OrderIDs: []uint64{},
	}
}

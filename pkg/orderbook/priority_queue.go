package orderbook

type PriorityQueue struct {
	isMaxHeap bool // Max => BuyOrder, Min => SellOrder
	levels    []*PriceLevel
}

func (pq *PriorityQueue) Len() int {
	return len(pq.levels)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	if pq.isMaxHeap {
		return pq.levels[i].Price > pq.levels[j].Price
	} else {
		return pq.levels[i].Price < pq.levels[j].Price
	}
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.levels[i], pq.levels[j] = pq.levels[j], pq.levels[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	pq.levels = append(pq.levels, x.(*PriceLevel))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.levels
	n := len(old)
	item := old[n-1]
	pq.levels = old[:n-1]
	return item
}

func NewPriorityQueue(isMaxHeap bool) *PriorityQueue {
	return &PriorityQueue{
		isMaxHeap: isMaxHeap,
		levels:    []*PriceLevel{},
	}
}

func (pq *PriorityQueue) GetLevels() []*PriceLevel {
	return pq.levels
}

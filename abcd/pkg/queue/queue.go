package queue

import (
	"container/heap"
	"log"
	"project/pkg/models"
)

type PriorityQueue []*models.Order

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Price > pq[j].Price
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*models.Order))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{}
	heap.Init(pq)
	return pq
}

func (pq *PriorityQueue) Enqueue(order *models.Order) {
	heap.Push(pq, order)
}

func (pq *PriorityQueue) Dequeue() *models.Order {
	if pq.Len() == 0 {
		log.Println("Queue is empty, nothing to dequeue")
		return nil
	}
	return heap.Pop(pq).(*models.Order)
}

func (pq *PriorityQueue) Peek() *models.Order {
	if pq.Len() == 0 {
		return nil
	}
	return (*pq)[0]
}

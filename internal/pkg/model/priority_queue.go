package model

type priorityQueue struct {
	items []*Item
}

func (q priorityQueue) Len() int {
	return len(q.items)
}

func (q priorityQueue) Less(i, j int) bool {
	if q.items[i].Expire <= 0 {
		return false
	}
	if q.items[j].Expire <= 0 {
		return true
	}
	return q.items[i].Expire < q.items[j].Expire
}

func (q priorityQueue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
	q.items[i].index = i
	q.items[j].index = j
}

func (q *priorityQueue) Push(x interface{}) {
	item := x.(*Item)
	item.index = len(q.items)
	q.items = append(q.items, item)
}

func (q *priorityQueue) Pop() interface{} {
	old := q.items
	n := len(old)
	item := old[n-1]
	item.index = -1
	q.items = old[0 : n-1]
	return item
}

func (q *priorityQueue) Top() *Item {
	if q.Len() == 0 {
		return nil
	}
	return q.items[0]
}

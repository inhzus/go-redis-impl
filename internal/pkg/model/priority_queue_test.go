package model

import (
	"container/heap"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQueuePush(t *testing.T) {
	queue := &priorityQueue{}
	for i := 0; i < 10; i++ {
		heap.Push(queue, &Item{})
	}
	assert.Equal(t, queue.Len(), 10, "Expected queue to have 10 elements")
}

func TestPriorityQueuePop(t *testing.T) {
	queue := &priorityQueue{}
	for i := 0; i < 10; i++ {
		heap.Push(queue, &Item{})
	}
	for i := 0; i < 5; i++ {
		item := heap.Pop(queue)
		assert.Equal(t, fmt.Sprintf("%T", item), "*model.Item", "Expected 'item' to be a '*model.Item'")
	}
	assert.Equal(t, queue.Len(), 5, "Expected queue to have 5 elements")
	for i := 0; i < 5; i++ {
		item := heap.Pop(queue)
		assert.Equal(t, fmt.Sprintf("%T", item), "*model.Item", "Expected 'item' to be a '*model.Item'")
	}
	assert.Equal(t, queue.Len(), 0, "Expected queue to have 0 elements")
	assert.Zero(t, len(queue.items), "Expected 'item' to be nil")
}

func TestPriorityQueueCheckOrder(t *testing.T) {
	queue := &priorityQueue{}
	for i := 10; i > 0; i-- {
		heap.Push(queue, newItem(fmt.Sprintf("key_%d", i), nil, time.Second*time.Duration(i)))
	}
	for i := 1; i <= 10; i++ {
		item := heap.Pop(queue).(*Item)
		assert.Equal(t, item.key, fmt.Sprintf("key_%d", i), "error")
	}
}

func TestPriorityQueueRemove(t *testing.T) {
	queue := &priorityQueue{}
	items := make(map[string]*Item)
	var itemRemove *Item
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key_%d", i)
		items[key] = newItem(key, nil, time.Duration(i)*time.Second)
		heap.Push(queue, items[key])

		if i == 2 {
			itemRemove = items[key]
		}
	}
	assert.Equal(t, queue.Len(), 5, "Expected queue to have 5 elements")
	if itemRemove == nil {
		return
	}
	heap.Remove(queue, itemRemove.index)
	assert.Equal(t, queue.Len(), 4, "Expected queue to have 4 elements")
	for {
		if len(queue.items) == 0 {
			break
		}
		item := heap.Pop(queue).(*Item)
		assert.NotEqual(t, itemRemove.key, item.key, "This element was not supposed to be in the queue")
	}

	assert.Equal(t, queue.Len(), 0, "The queue is supposed to be with 0 items")
}

func TestPriorityQueueUpdate(t *testing.T) {
	queue := &priorityQueue{}
	item := &Item{}
	heap.Push(queue, item)
	assert.Equal(t, queue.Len(), 1, "The queue is supposed to be with 1 item")

	item.key = "newKey"
	heap.Fix(queue, item.index)
	newItem := heap.Pop(queue).(*Item)
	assert.Equal(t, newItem.key, "newKey", "The item key didn't change")
	assert.Equal(t, queue.Len(), 0, "The queue is supposed to be with 0 items")
}

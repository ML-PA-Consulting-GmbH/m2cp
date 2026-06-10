package amqp

import (
	"container/list"
	"m2cp/networks/stats"
	"sync"
)

type sendJobQueue struct {
	items *list.List
	stats *stats.QueueStats
	lock  sync.Mutex
}

func newSendJobQueue(stats *stats.QueueStats) sendJobQueue {
	return sendJobQueue{
		items: list.New(),
		stats: stats,
	}
}

func (q *sendJobQueue) Enqueue(item *SendJob) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items.PushBack(item)
	q.stats.Len.Set(uint64(q.items.Len()))
}

func (q *sendJobQueue) Dequeue() *SendJob {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.items.Len() == 0 {
		return nil
	}
	item := q.items.Front()
	q.items.Remove(item)
	q.stats.Len.Set(uint64(q.items.Len()))
	return item.Value.(*SendJob)
}

func (q *sendJobQueue) Peek() *SendJob {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.items.Len() == 0 {
		return nil
	}
	item := q.items.Front()
	return item.Value.(*SendJob)
}

func (q *sendJobQueue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.items.Len()
}

func (q *sendJobQueue) Clear() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items.Init()
	q.stats.Len.Set(uint64(q.items.Len()))
}

func (q *sendJobQueue) DequeueEnd() *SendJob {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.items.Len() == 0 {
		return nil
	}
	item := q.items.Back()
	q.items.Remove(item)
	q.stats.Len.Set(uint64(q.items.Len()))
	return item.Value.(*SendJob)
}

func (q *sendJobQueue) EnqueueFront(item *SendJob) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items.PushFront(item)
	q.stats.Len.Set(uint64(q.items.Len()))
}

func (q *sendJobQueue) RemoveJobWithDeliveryTag(tag uint64) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	for e := q.items.Front(); e != nil; e = e.Next() {
		if e.Value.(*SendJob).DeliveryTag == tag {
			q.items.Remove(e)
			q.stats.Len.Set(uint64(q.items.Len()))
			return nil
		}
	}
	return nil
}

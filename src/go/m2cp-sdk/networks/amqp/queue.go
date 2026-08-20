package amqp

import (
	"container/list"
	"context"
	"errors"
	"m2cp/networks/stats"
	"runtime/trace"
	"sync"
	"time"
)

type sendJobQueue struct {
	ctx   context.Context
	name  string
	items *list.List
	stats *stats.QueueStats
	lock  sync.Mutex
}

func newSendJobQueue(ctx context.Context, stats *stats.QueueStats, name string) sendJobQueue {
	return sendJobQueue{
		items: list.New(),
		stats: stats,
		ctx:   ctx,
		name:  name,
	}
}

func (q *sendJobQueue) Enqueue(item *SendJob) {
	lockTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.Enqueue.Lock")
	q.lock.Lock()
	defer q.lock.Unlock()
	lockTrace.End()
	enqTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.Enqueue.Work")
	q.items.PushBack(item)
	q.stats.Len.Set(uint64(q.items.Len()))
	item.EnqueueTime = time.Now()
	enqTrace.End()
}

func (q *sendJobQueue) Dequeue() *SendJob {
	lockTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.Dequeue.Lock")
	q.lock.Lock()
	defer q.lock.Unlock()
	lockTrace.End()

	if q.items.Len() == 0 {
		return nil
	}
	deqTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.Dequeue.Work")
	item := q.items.Front()
	q.items.Remove(item)
	q.stats.Len.Set(uint64(q.items.Len()))
	job := item.Value.(*SendJob)
	q.stats.Delta.Add(time.Now().Sub(job.EnqueueTime))
	deqTrace.End()
	return job
}

func (q *sendJobQueue) Peek() *SendJob {
	lockTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.Peek.Lock")
	q.lock.Lock()
	defer q.lock.Unlock()
	lockTrace.End()

	if q.items.Len() == 0 {
		return nil
	}
	item := q.items.Front()
	return item.Value.(*SendJob)
}

func (q *sendJobQueue) Len() int {
	lockTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.Len.Lock")
	q.lock.Lock()
	defer q.lock.Unlock()
	lockTrace.End()
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
	lockTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.RemoveJobWithDeliveryTag.Lock")
	q.lock.Lock()
	defer q.lock.Unlock()
	lockTrace.End()
	workTrace := trace.StartRegion(q.ctx, q.name+".sendJobQueue.RemoveJobWithDeliveryTag.Work")
	for e := q.items.Front(); e != nil; e = e.Next() {
		if e.Value.(*SendJob).deferredConfirm.DeliveryTag == tag {
			q.items.Remove(e)
			q.stats.Len.Set(uint64(q.items.Len()))
			q.stats.Delta.Add(time.Since(e.Value.(*SendJob).EnqueueTime))
			workTrace.End()
			return nil
		}
	}
	workTrace.End()
	return errors.New("job with delivery tag not found")
}

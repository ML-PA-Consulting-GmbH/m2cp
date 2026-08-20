package stats

import (
	"m2cp"
	"sync"
	"time"
)

type NetworkStats struct {
	Sender SenderStats
	// Receiver                                                     WorkerStats
	lockMessagesSent, lockMessagesSentAck, lockMessagesSentError sync.Mutex
}

func (o *NetworkStats) GetSenderStats() m2cp.SenderStats {
	return &o.Sender
}

type Counter struct {
	counter uint64
	lock    sync.Mutex
}

func (c *Counter) Inc() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.counter = c.counter + 1
}

func (c *Counter) Get() uint64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.counter
}

func (c *Counter) Set(value uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.counter = value
}

func (c *Counter) Sum(value uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.counter += value
}

func (c *Counter) SetIfMax(value uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if value > c.counter {
		c.counter = value
	}
}

func (c *Counter) SetIfMin(value uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if (c.counter == 0 || value < c.counter) && value > 0 {
		c.counter = value
	}
}

type SenderStats struct {
	Transmitted Counter
	Acks        Counter
	Errors      Counter
	QueueSend   QueueStats
	QueueAck    QueueStats
	Connected   bool
}

func (o *SenderStats) GetConnected() bool {
	return o.Connected
}

func (o *SenderStats) GetTransmitted() uint64 {
	return o.Transmitted.Get()
}

func (o *SenderStats) GetAcked() uint64 {
	return o.Acks.Get()
}

func (o *SenderStats) GetErrors() uint64 {
	return o.Errors.Get()
}

func (o *SenderStats) GetQueueSendLen() uint64 {
	return o.QueueSend.Len.Get()
}
func (o *SenderStats) GetQueueSendMax() uint64 {
	return o.QueueSend.Max.Get()
}
func (o *SenderStats) GetQueueAckLen() uint64 {
	return o.QueueAck.Len.Get()
}
func (o *SenderStats) GetQueueAckMax() uint64 {
	return o.QueueAck.Max.Get()
}

type QueueStats struct {
	Len   Counter
	Max   Counter
	Delta DeltaTimes
}

type DeltaTimes struct {
	Min   Counter
	Max   Counter
	Sum   Counter
	Count Counter
	Avg   Counter
}

func (o *DeltaTimes) Add(delta time.Duration) {
	v := uint64(delta.Milliseconds())
	o.Count.Inc()
	o.Sum.Sum(v)
	o.Min.SetIfMin(v)
	o.Max.SetIfMax(v)
	o.Avg.Set(o.calcAvg())
}

func (o *DeltaTimes) calcAvg() uint64 {
	count := o.Count.Get()
	if count == 0 {
		return 0
	}
	return o.Sum.Get() / count
}

// Package ratelimit RateLimit Algorithm Based on Token Bucket
package flowctrl

import (
	"io"
	"sync"
	"time"
)

const (
	tokenWindow  = 50 * time.Millisecond
	minWindowCap = 64
)

type Controller struct {
	capacity      int
	threshold     int
	cond          *sync.Cond
	done          chan struct{}
	ratePerSecond int
	closeOnce     *sync.Once
}

func NewController(ratePerSecond int) *Controller {
	capacity := ratePerSecond * int(tokenWindow) / int(time.Second)
	if capacity < minWindowCap {
		capacity = minWindowCap
	}
	c := &Controller{
		ratePerSecond: ratePerSecond,
		threshold:     capacity,
		capacity:      capacity,
		cond:          sync.NewCond(new(sync.Mutex)),
		done:          make(chan struct{}, 1),
		closeOnce:     &sync.Once{},
	}
	// produce token asynchronously
	go c.run(capacity)
	return c
}

func (c *Controller) acquire(size int) int {
	c.cond.L.Lock()
	for c.capacity == 0 {
		c.cond.Wait()
	}
	if size > c.capacity {
		size = c.capacity
	}
	c.capacity -= size
	c.cond.L.Unlock()
	return size
}

func (c *Controller) fill(size int) {
	if size <= 0 {
		return
	}
	c.cond.L.Lock()
	c.capacity += size
	if c.capacity > c.threshold {
		c.capacity = c.threshold
	}
	c.cond.L.Unlock()
	c.cond.Broadcast()
}

func (c *Controller) run(capacity int) {
	t := time.NewTicker(tokenWindow)
	for {
		select {
		case <-t.C:
			c.cond.L.Lock()
			c.capacity = capacity
			c.cond.L.Unlock()
			c.cond.Broadcast()
		case <-c.done:
			t.Stop()
			return
		}
	}
}

// Close release token-producing goroutine
func (c *Controller) Close() error {
	c.closeOnce.Do(func() {
		c.done <- struct{}{}
	})
	return nil
}

func (c *Controller) Reader(underlying io.Reader) io.Reader {
	return &rateReader{
		underlying: underlying,
		c:          c,
	}
}

func (c *Controller) Writer(underlying io.Writer) io.Writer {
	return &rateWriter{
		underlying: underlying,
		c:          c,
	}
}

func (c *Controller) GetRateLimit() int {
	return c.ratePerSecond
}

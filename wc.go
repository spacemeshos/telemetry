package telemetry

import (
	"math"
	"sync"
	"sync/atomic"
)

type WC struct {
	Value uint64
	cond  sync.Cond
	mu    sync.Mutex
}

func (c *WC) Wait(index uint64) (r bool) {
	r = true
	if atomic.LoadUint64(&c.Value) == index {
		// mostly when executes consequentially
		return
	}
	c.mu.Lock()
	if c.cond.L == nil {
		c.cond.L = &c.mu
	}
	for c.Value != index {
		if c.Value > index {
			if c.Value == math.MaxUint64 {
				r = false
				break
			}
			panic("index continuity is broken")
		}
		c.cond.Wait()
	}
	c.mu.Unlock()
	return
}

func (c *WC) Inc() (r bool) {
	c.mu.Lock()
	if c.cond.L == nil {
		c.cond.L = &c.mu
	}
	if c.Value < math.MaxUint64 {
		atomic.AddUint64(&c.Value, 1)
		r = true
	}
	c.mu.Unlock()
	if r {
		c.cond.Broadcast()
	}
	return
}

func (c *WC) Set(v uint64) (r bool) {
	c.mu.Lock()
	if c.cond.L == nil {
		c.cond.L = &c.mu
	}
	if c.Value < math.MaxUint64 || c.Value < v {
		atomic.StoreUint64(&c.Value, v)
		r = true
	}
	c.mu.Unlock()
	if r {
		c.cond.Broadcast()
	}
	return
}

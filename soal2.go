package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type Pool[T any] interface {
	Get() *T
	Put(*T)
	Stats() PoolStats
	Close() error
}

type PoolStats struct {
	TotalAllocations int64
	CurrentPoolSize  int64
	HitRatio         float64
}

type pool[T any] struct {
	newFunc func() *T
	mu      sync.Mutex
	cond    *sync.Cond
	closed  bool
	items   []*T

	totalAllocations int64
	hits             int64
	misses           int64
}

func NewPool[T any](newFunc func() *T, initialSize int) *pool[T] {
	p := &pool[T]{newFunc: newFunc, items: make([]*T, 0, initialSize)}

	p.cond = sync.NewCond(&p.mu)

	for i := 0; i < initialSize; i++ {
		p.items = append(p.items, newFunc())
		atomic.AddInt64(&p.totalAllocations, 1)
	}

	return p
}

func (p *pool[T]) Get() *T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.items) == 0 && !p.closed {
		obj := p.newFunc()
		atomic.AddInt64(&p.totalAllocations, 1)
		atomic.AddInt64(&p.misses, 1)
		return obj

	}

	if p.closed {
		return nil
	}

	n := len(p.items) - 1
	obj := p.items[n]
	p.items = p.items[:n]

	atomic.AddInt64(&p.hits, 1)
	return obj
}

func (p *pool[T]) Stats() PoolStats {
	total := atomic.LoadInt64(&p.totalAllocations)
	hits := atomic.LoadInt64(&p.hits)
	miss := atomic.LoadInt64(&p.misses)

	hitRatio := 0.0

	totalA := hits + miss

	if totalA > 0 {
		hitRatio = float64(hits) / float64(totalA)
		// missRatio = float64(miss) / float64(totalA)
	}

	return PoolStats{
		TotalAllocations: total,
		CurrentPoolSize:  int64(len(p.items)),
		HitRatio:         hitRatio,
	}
}

func (p *pool[T]) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errors.New("pool closed")
	}

	p.closed = true
	p.items = nil
	p.cond.Broadcast()
	return nil
}

func (p *pool[T]) Put(obj *T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.items = append(p.items, obj)
	p.cond.Signal()
}

func Soal2() {

	p := NewPool(func() *int { return new(int) }, 2)

	obj1 := p.Get()
	*obj1 = 42
	fmt.Println("Get:", *obj1)

	p.Put(obj1)

	obj2 := p.Get()
	fmt.Println("Get Again:", *obj2)

	fmt.Printf("Stats: %+v\n", p.Stats())

}

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Connection interface {
	ID() string
}

type DBPool interface {
	GetConnection(ctx context.Context) (Connection, error)
	ReturnConnection(conn Connection) error
	HealthCheck() error
	Stats() PoolStats
	Close() error
}
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

type DummyConnection struct {
	id        string
	createdAt time.Time
}

func NewDummyConnection(id string) *DummyConnection {
	return &DummyConnection{
		id: id,
	}
}

type dbPool struct {
	conns chan Connection

	mu sync.Mutex

	newConnFn func(ctx context.Context)
}

type DBPoolConfig struct {
	MaxConns int
}

func NewDBPool() (*dbPool, error) {
	p := &dbPool{
		conns: make(chan Connection),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	conn, err := p.newConnWithRetry(ctx)
	cancel()
	if err == nil {
		p.conns <- conn
	}

	return p, nil
}

func (p *dbPool) newConnWithRetry(ctx context.Context) (Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, err := p.newConnFn(ctx)
	if err == nil {
		return conn, nil
	}

	return nil, nil
}

func Soal4() {
	pool, err := NewDBPool()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pool)
}

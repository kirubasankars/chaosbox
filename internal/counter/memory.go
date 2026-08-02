package counter

import (
	"context"
	"sync/atomic"
)

// Memory is a process-local Counter backed by an atomic int64.
type Memory struct {
	n atomic.Int64
}

func NewMemory() *Memory {
	return &Memory{}
}

func (c *Memory) Incr(ctx context.Context) (int64, error) {
	return c.n.Add(1), nil
}

func (c *Memory) Decr(ctx context.Context) (int64, error) {
	return c.n.Add(-1), nil
}

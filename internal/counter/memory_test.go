package counter_test

import (
	"context"
	"testing"

	"chaosbox/internal/counter"
)

func TestMemoryCounter_GetIncrDecr(t *testing.T) {
	c := counter.NewMemory()
	ctx := context.Background()

	if n, err := c.Get(ctx); err != nil || n != 0 {
		t.Fatalf("Get = %d, %v; want 0, nil", n, err)
	}
	if n, err := c.Incr(ctx); err != nil || n != 1 {
		t.Fatalf("Incr = %d, %v; want 1, nil", n, err)
	}
	if n, err := c.Get(ctx); err != nil || n != 1 {
		t.Fatalf("Get = %d, %v; want 1, nil", n, err)
	}
	if n, err := c.Decr(ctx); err != nil || n != 0 {
		t.Fatalf("Decr = %d, %v; want 0, nil", n, err)
	}
}

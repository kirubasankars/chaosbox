package counter_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"

	"chaosbox/internal/counter"
)

// redisDSN returns a DSN to test against. If CHAOSBOX_TEST_REDIS_DSN is set, it
// is used as-is (pointing at a real Redis instance, e.g. in CI). Otherwise
// an in-process miniredis server is started and torn down automatically.
func redisDSN(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("CHAOSBOX_TEST_REDIS_DSN"); dsn != "" {
		return dsn
	}
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	return "redis://" + mr.Addr() + "/0"
}

func TestRedisCounter_IncrDecr(t *testing.T) {
	c, err := counter.NewRedis(redisDSN(t))
	if err != nil {
		t.Fatalf("NewRedis: %v", err)
	}
	ctx := context.Background()

	// Use deltas rather than absolute values so this also passes against a
	// real, possibly non-empty Redis instance (CHAOSBOX_TEST_REDIS_DSN).
	if n, err := c.Get(ctx); err != nil {
		t.Fatalf("Get (missing key): %v", err)
	} else if n != 0 && os.Getenv("CHAOSBOX_TEST_REDIS_DSN") == "" {
		// miniredis starts empty; real Redis may already have a value.
		t.Fatalf("Get (missing key) = %d; want 0", n)
	}

	base, err := c.Incr(ctx)
	if err != nil {
		t.Fatalf("Incr: %v", err)
	}
	if n, err := c.Get(ctx); err != nil || n != base {
		t.Fatalf("Get = %d, %v; want %d, nil", n, err, base)
	}
	if n, err := c.Incr(ctx); err != nil || n != base+1 {
		t.Fatalf("Incr = %d, %v; want %d, nil", n, err, base+1)
	}
	if n, err := c.Decr(ctx); err != nil || n != base {
		t.Fatalf("Decr = %d, %v; want %d, nil", n, err, base)
	}
	if n, err := c.Decr(ctx); err != nil || n != base-1 {
		t.Fatalf("Decr = %d, %v; want %d, nil", n, err, base-1)
	}
}

func TestRedisCounter_SharedAcrossInstances(t *testing.T) {
	dsn := redisDSN(t)

	// Two independent Counter instances pointed at the same Redis DSN
	// simulate two chaosbox nodes sharing a counter backend.
	a, err := counter.NewRedis(dsn)
	if err != nil {
		t.Fatalf("NewRedis (a): %v", err)
	}
	b, err := counter.NewRedis(dsn)
	if err != nil {
		t.Fatalf("NewRedis (b): %v", err)
	}
	ctx := context.Background()

	base, err := a.Incr(ctx)
	if err != nil {
		t.Fatalf("a.Incr: %v", err)
	}
	if n, err := b.Incr(ctx); err != nil || n != base+1 {
		t.Fatalf("b.Incr = %d, %v; want %d, nil", n, err, base+1)
	}
	if n, err := a.Decr(ctx); err != nil || n != base {
		t.Fatalf("a.Decr = %d, %v; want %d, nil", n, err, base)
	}
}

func TestNewRedis_UnreachableFailsFast(t *testing.T) {
	// Bind then immediately close a real port, so connecting to it fails
	// fast with "connection refused" instead of retrying/backing off.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	if _, err := counter.NewRedis("redis://" + addr + "/0"); err == nil {
		t.Fatal("NewRedis: expected error for unreachable DSN, got nil")
	}
}

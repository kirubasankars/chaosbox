// Package counter implements the /count/incr and /count/decr backends:
// an in-memory atomic counter, or a shared Redis-backed counter.
package counter

import "context"

type Counter interface {
	Incr(ctx context.Context) (int64, error)
	Decr(ctx context.Context) (int64, error)
}

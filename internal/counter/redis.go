package counter

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// Redis is a Counter backed by a shared Redis key, letting multiple chaosbox
// instances see the same value.
type Redis struct {
	client *redis.Client
	key    string
}

func NewRedis(dsn string) (*Redis, error) {
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Redis{client: client, key: "chaosbox:count"}, nil
}

func (c *Redis) Get(ctx context.Context) (int64, error) {
	n, err := c.client.Get(ctx, c.key).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return n, err
}

func (c *Redis) Incr(ctx context.Context) (int64, error) {
	return c.client.Incr(ctx, c.key).Result()
}

func (c *Redis) Decr(ctx context.Context) (int64, error) {
	return c.client.Decr(ctx, c.key).Result()
}

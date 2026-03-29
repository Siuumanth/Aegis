package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// tis is liek a repository
type Backend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// RunScript(ctx context.Context, script *Script, keys []string, args ...any) (any, error)
}

type Client struct {
	rdb *goredis.Client
}

func NewClient(addr string) *Client {
	return &Client{
		rdb: goredis.NewClient(&goredis.Options{Addr: addr}),
	}
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	//.Println("SETing key", key, value, ttl)
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// lua script for invalidating a tag
// delete forward index and reverse index
var invalidateScript = goredis.NewScript(`
local keys = redis.call('SMEMBERS', KEYS[1])  -- get all keys under tag
if #keys > 0 then
    for _, key in ipairs(keys) do
        redis.call('SREM', 'key-tags:' .. key, KEYS[2])  -- remove tag from reverse index
    end
    redis.call('DEL', unpack(keys))  -- delete the actual keys
end
redis.call('DEL', KEYS[1])  -- delete forward index
return #keys
`)

// tis is liek a repository
type Backend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	// increase TTL
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// tag operations
	AddKeyToSet(ctx context.Context, set string, member string) error
	RemoveKeyFromSet(ctx context.Context, set string, member string) error
	GetSetMembers(ctx context.Context, set string) ([]string, error)
	DeleteKeyTags(ctx context.Context, key string, tags []string) error
	InvalidateTag(ctx context.Context, tagKey string, tag string) (int64, error)
}

type Client struct {
	rdb *goredis.Client
}

func NewClient(addr string) *Client {
	return &Client{
		rdb: goredis.NewClient(&goredis.Options{
			Addr:     addr,
			Protocol: 2,
		}),
	}
}

// InvalidateTag atomically deletes all keys under a tag via Lua script
func (c *Client) InvalidateTag(ctx context.Context, tagKey string, tag string) (int64, error) {
	result, err := invalidateScript.Run(ctx, c.rdb, []string{tagKey, tag}).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// tag operations

func (c *Client) AddKeyToSet(ctx context.Context, set string, member string) error {
	return c.rdb.SAdd(ctx, set, member).Err()
}

func (c *Client) RemoveKeyFromSet(ctx context.Context, set string, member string) error {
	return c.rdb.SRem(ctx, set, member).Err()
}

func (c *Client) GetSetMembers(ctx context.Context, set string) ([]string, error) {
	return c.rdb.SMembers(ctx, set).Result()
}

// a pipeline to deleete all tags for a keyy using a single script
func (c *Client) DeleteKeyTags(ctx context.Context, key string, tags []string) error {
	pipe := c.rdb.Pipeline()
	// Creates a Redis pipeline
	// to batch multiple commands into ONE network round trip
	for _, tag := range tags {
		pipe.SRem(ctx, tag, key)
	}
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	return err
}

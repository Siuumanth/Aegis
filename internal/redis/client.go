package redis

import (
	"Aegis/internal/resp"
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// InvalidateTag atomically deletes all keys under a tag and cleans up all associated indexes.
// KEYS[1] = The Forward Index Key (e.g., "tag:profile")
// KEYS[2] = The Raw Tag Name (e.g., "profile")
var invalidateScript = goredis.NewScript(`
-- KEYS[1] is "tag:profile"
-- KEYS[2] is "profile" (the raw tag name)
local target_tag_set = KEYS[1]

-- 1. Get all user keys like "user:0", "user:1"
local keys = redis.call('SMEMBERS', target_tag_set)

if #keys > 0 then
    for _, key in ipairs(keys) do
        -- 2. Reverse Index key name
        local rev_idx = 'key-tags:' .. key
        
        -- 3. Get the raw tag names from the screenshot (["users", "profile"])
        local associated_tags = redis.call('SMEMBERS', rev_idx)
        
        for _, t_name in ipairs(associated_tags) do
            -- 4. CONSTRUCT THE FORWARD KEY
            -- Since your screenshot shows "users", we make "tags:users"
            local forward_idx_key = "tag:" .. t_name
            
            -- 5. YANK KEY FROM ALL OTHER SETS
            -- This is what removes "user:0" from "tag:users"
            redis.call('SREM', forward_idx_key, key)
        end
        
        -- 6. DELETE METADATA & ACTUAL DATA
        redis.call('DEL', rev_idx)
        redis.call('DEL', key)
    end
end

-- 7. DELETE THE PRIMARY TARGET SET
redis.call('DEL', target_tag_set)

return #keys
`)

// tis is liek a repository
type Backend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	PassThrough(ctx context.Context, cmd *resp.Command) (any, error)

	// increase TTL
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// tag operations
	AddKeyToSet(ctx context.Context, set string, member string) error
	RemoveKeyFromSet(ctx context.Context, set string, member string) error
	GetSetMembers(ctx context.Context, set string) ([]string, error)
	DeleteKeyTags(ctx context.Context, key string, revKey string, tags []string) error
	InvalidateTag(ctx context.Context, tagKey string, tag string) (int64, error)

	// Pipe
	StartPipeline(ctx context.Context) goredis.Pipeliner
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

// pass raw bytes to redis
// raw bytes not possible thru redis client so rebuild cmd and send
func (c *Client) PassThrough(ctx context.Context, cmd *resp.Command) (any, error) {
	args := make([]any, 0, len(cmd.Args)+2)
	args = append(args, cmd.Name)
	if cmd.Key != "" {
		args = append(args, cmd.Key)
	}
	for _, arg := range cmd.Args {
		args = append(args, arg)
	}
	return c.rdb.Do(ctx, args...).Result()
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
func (c *Client) DeleteKeyTags(ctx context.Context, key string, revKey string, tags []string) error {
	pipe := c.rdb.Pipeline()
	// Creates a Redis pipeline
	// to batch multiple commands into ONE network round trip
	for _, tag := range tags {
		pipe.SRem(ctx, tag, key)
	}
	pipe.Del(ctx, key)
	pipe.Del(ctx, revKey)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *Client) StartPipeline(ctx context.Context) goredis.Pipeliner {
	return c.rdb.Pipeline()
}

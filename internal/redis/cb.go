package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
	gobreaker "github.com/sony/gobreaker/v2"

	"Aegis/config"
	"Aegis/internal/resp"
	"Aegis/internal/shared"
)

type CBBackend struct {
	inner   Backend
	breaker *gobreaker.CircuitBreaker[any]
}

func NewCBBackend(inner Backend, cfg *config.RedisConfig) *CBBackend {
	settings := gobreaker.Settings{
		Name:        "redis",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     4 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("[CB] state change: %s → %s\n", from.String(), to.String())
		},
	}
	return &CBBackend{
		inner:   inner,
		breaker: gobreaker.NewCircuitBreaker[any](settings),
	}
}

// global fucnton executor
func (c *CBBackend) exec(fn func() (any, error)) (any, error) {
	result, err := c.breaker.Execute(fn)
	if err != nil {
		if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
			log.Printf("[CB] circuit open — rejecting request\n")
		} else {
			log.Printf("[CB] redis error: %v\n", err)
		}
		return nil, err
	}
	return result, nil
}

func (c *CBBackend) Ping(ctx context.Context) error {
	_, err := c.exec(func() (any, error) {
		return nil, c.inner.Ping(ctx)
	})
	if err != nil {
		log.Printf("PING failed: \n")
		return err
	}
	return err
}

// custom handler on each command

func (c *CBBackend) Get(ctx context.Context, key string) (string, error) {
	result, err := c.exec(func() (any, error) {
		val, err := c.inner.Get(ctx, key)
		if err == shared.ErrGoRedisNil {
			return "", shared.ErrGoRedisNil // don't trip CB on cache miss
		}
		return val, err
	})
	if err != nil {
		if err == shared.ErrGoRedisNil {
			return "", err
		}
		if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
			log.Printf("[CB] circuit open — GET %s rejected\n", key)
		} else {
			log.Printf("[CB] GET %s failed: %v\n", key, err)
		}
		return "", err
	}
	val, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("CB: expected []string, got %T", result)
	}
	return val, nil
}

func (c *CBBackend) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	_, err := c.exec(func() (any, error) {
		return nil, c.inner.Set(ctx, key, value, ttl)
	})
	if err != nil {
		log.Printf("[CB] SET %s failed: %v\n", key, err)
	}
	return err
}

func (c *CBBackend) Del(ctx context.Context, keys ...string) error {
	_, err := c.exec(func() (any, error) {
		return nil, c.inner.Del(ctx, keys...)
	})
	if err != nil {
		log.Printf("[CB] DEL %v failed: %v\n", keys, err)
	}
	return err
}

func (c *CBBackend) PassThrough(ctx context.Context, cmd *resp.Command) (any, error) {
	result, err := c.exec(func() (any, error) {
		return c.inner.PassThrough(ctx, cmd)
	})
	if err != nil {
		log.Printf("[CB] PassThrough %s failed: %v\n", cmd.Name, err)
		return nil, err
	}
	return result, nil
}

func (c *CBBackend) Expire(ctx context.Context, key string, ttl time.Duration) error {
	_, err := c.exec(func() (any, error) {
		return nil, c.inner.Expire(ctx, key, ttl)
	})
	if err != nil {
		log.Printf("[CB] EXPIRE %s failed: %v\n", key, err)
	}
	return err
}

func (c *CBBackend) AddKeyToSet(ctx context.Context, set string, member string) error {
	_, err := c.exec(func() (any, error) {
		return nil, c.inner.AddKeyToSet(ctx, set, member)
	})
	if err != nil {
		log.Printf("[CB] AddKeyToSet set=%s member=%s failed: %v\n", set, member, err)
	}
	return err
}

func (c *CBBackend) RemoveKeyFromSet(ctx context.Context, set string, member string) error {
	_, err := c.exec(func() (any, error) {
		return nil, c.inner.RemoveKeyFromSet(ctx, set, member)
	})
	if err != nil {
		log.Printf("[CB] RemoveKeyFromSet set=%s member=%s failed: %v\n", set, member, err)
	}
	return err
}

func (c *CBBackend) GetSetMembers(ctx context.Context, set string) ([]string, error) {
	result, err := c.exec(func() (any, error) {
		return c.inner.GetSetMembers(ctx, set)
	})
	if err != nil {
		log.Printf("[CB] GetSetMembers set=%s failed: %v\n", set, err)
		return nil, err
	}
	val, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("CB: expected []string, got %T", result)
	}
	return val, nil
}

func (c *CBBackend) DeleteKeyTags(ctx context.Context, key string, revKey string, tags []string) error {
	_, err := c.exec(func() (any, error) {
		return nil, c.inner.DeleteKeyTags(ctx, key, revKey, tags)
	})
	if err != nil {
		log.Printf("[CB] DeleteKeyTags key=%s failed: %v\n", key, err)
	}
	return err
}

func (c *CBBackend) InvalidateTag(ctx context.Context, tagKey string, tag string) (int64, error) {
	result, err := c.exec(func() (any, error) {
		return c.inner.InvalidateTag(ctx, tagKey, tag)
	})
	if err != nil {
		log.Printf("[CB] InvalidateTag tag=%s failed: %v\n", tag, err)
		return 0, err
	}
	val, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("CB: expected []string, got %T", result) // idk if tis right
	}
	return val, nil
}

func (c *CBBackend) StartPipeline(ctx context.Context) goredis.Pipeliner {
	// pipeline can't go through CB cleanly — pass through directly
	// individual pipeline commands will fail naturally if redis is down
	return c.inner.StartPipeline(ctx)
}

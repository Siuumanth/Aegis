package tags

import (
	"Aegis/internal/redis"
	"context"
	"fmt"
)

/*
PLAN:

policies:
  - name: "user-profiles"
    match:
      pattern: "user:*"
    config:
      ttl: 45s
      min_ttl: 5s
      max_ttl: 10m
      singleflight: true
      tags: [users,profile]

whenever set req recieved, an event is pushed to the tagService
- tis will talk to redis and add that key to the set of the tags
func (c *Client) AddKeyToSet(ctx context.Context, tag string, key string) error

- and also there is a tag based invalidatin function, so need to write script for that as well

*/

type TagEvent struct {
	key  string
	tags []string
}

// contains channels for async
type TagService struct {
	redis        redis.Backend
	registerChan chan TagEvent
	deleteChan   chan string
}

func NewTagService(redisClient redis.Backend, bufSize int) *TagService {
	return &TagService{
		redis:        redisClient,
		registerChan: make(chan TagEvent, bufSize),
		deleteChan:   make(chan string, bufSize),
	}
}

// Start spawns N workers draining both channels
func (t *TagService) Start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case ev := <-t.registerChan:
					t.register(ctx, ev)
				case key := <-t.deleteChan:
					t.delete(ctx, key)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Register enqueues a tag registration event from SET or ATAG, non-blocking
// Register parses raw args for AEGIS.TAG modifiers and enqueues tag event
func (t *TagService) Register(key string, policyTags []string, args []string) {
	mode, extraTags := parseTagArgs(args)

	if mode == TagModeSkip {
		return
	}

	var all []string
	switch mode {
	case TagModeOverride:
		all = extraTags // ignore policy tags if Aegis.TagOnly
	case TagModeNormal:
		all = append(policyTags, extraTags...)
	}

	if len(all) == 0 {
		return
	}

	select {
	case t.registerChan <- TagEvent{key: key, tags: all}:
	default: // drop silently v1
	}
}

// Delete enqueues a delete event, non-blocking
func (t *TagService) Delete(key string) {
	select {
	case t.deleteChan <- key:
	default: // drop silently for v1
	}
}

func (t *TagService) register(ctx context.Context, ev TagEvent) {
	// 1. Create the pipeline
	pipe := t.redis.StartPipeline(ctx)

	for _, tag := range ev.tags {
		tKey := tagKey(tag)
		rKey := reverseKey(ev.key)

		// 2. Queue the commands (No network call yet)
		// Forward index: tag -> key
		pipe.SAdd(ctx, tKey, ev.key)
		// Reverse index: key -> tag
		pipe.SAdd(ctx, rKey, tag)
	}

	// 3. Execute all at once (One network round-trip)
	_, err := pipe.Exec(ctx)
	if err != nil {
		// Log it, but don't crash. Since it's async,
		// the client is already gone anyway.
		fmt.Printf("[TagService] Pipeline failed for key %s: %v\n", ev.key, err)
	}
}

func (t *TagService) delete(ctx context.Context, key string) {
	tags, err := t.redis.GetSetMembers(ctx, reverseKey(key))
	if err != nil {
		return
	}

	// build tag keys for forward index cleanup
	tagKeys := make([]string, len(tags))
	for i, tag := range tags {
		tagKeys[i] = tagKey(tag)
	}

	// pipeline: SREM from all forward indexes + DEL reverse index
	t.redis.DeleteKeyTags(ctx, key, reverseKey(key), tagKeys)
}

// delete all keys under a tag via Lua script
func (t *TagService) Invalidate(ctx context.Context, tag string) (int64, error) {
	return t.redis.InvalidateTag(ctx, tagKey(tag), tag)
}

// get tag key name
func tagKey(tag string) string {
	return fmt.Sprintf("tag:%s", tag)
}

// get key -> tags set name
func reverseKey(key string) string {
	return fmt.Sprintf("key-tags:%s", key)
}

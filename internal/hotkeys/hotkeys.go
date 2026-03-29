package hotkeys

import (
	"Aegis/config"
	"Aegis/internal/redis"
	"context"
	"sync"
	"time"
)

/*
PLAN:
- hot keys increment and updates are done in the background, async
2 options to increase TTL:
store last Seen in map
1. Get curent ttl, multiply it (but can cause too many round trips )
and thers no use if 1 seconds increased to 2...
2. multiply with policy ttl * multiplier
- better but too much frequent can cause capping

FINAL SOLUTION: BEST

- hot key map is a map to store key:count, last increased
- after a lot of painful thinking, last increasd seemd like the best approach
- we can even track to not increase keys too frequently in order to avoid hammering redis

STATE:
- state shud know the policyconfig, max keys , cleanup interval n stuff

FUNCTIONS NEEDED:
- worker pool running to get event and process it

ukw ill just cleanup when
policy * multiplier + last updated is more than the current time , then ill know th key has expired

*/

type HotKeyEntry struct {
	Count         int64
	LastIncreased time.Time // last time the hot key was multiplied
}

type hkEvent struct {
	key    string
	policy *config.PolicyConfig
}

type HotKeyService struct {
	mu                sync.RWMutex
	m                 map[string]*HotKeyEntry // map to store key:[count, last increased]
	hkChan            chan hkEvent            // channel to get events
	maxKeys           int
	cleanup           time.Duration
	redis             redis.Backend
	minExtendInterval time.Duration
}

// TODO: add tis to yaml

func NewHotKeyService(maxKeys int, bufSize int, redisClient redis.Backend) *HotKeyService {
	return &HotKeyService{
		m:       make(map[string]*HotKeyEntry),
		hkChan:  make(chan hkEvent, bufSize),
		maxKeys: maxKeys,
		redis:   redisClient,
	}
}

// init functoin , spawn a smallworker pool
// Start spawns N workers draining the event channel
func (h *HotKeyService) Start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case ev := <-h.hkChan:
					h.increment(ctx, ev.key, ev.policy)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Track enqueues a key event, non-blocking
func (h *HotKeyService) Track(key string, policy *config.PolicyConfig) {
	select {
	case h.hkChan <- hkEvent{key: key, policy: policy}:
	default: // channel full, drop silently
	}
}

func (h *HotKeyService) increment(ctx context.Context, key string, policy *config.PolicyConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// safe incrememnnt

	entry, ok := h.m[key]
	if !ok {
		if len(h.m) >= h.maxKeys {
			return
		}
		h.m[key] = &HotKeyEntry{Count: 1}
		return
	}

	entry.Count++

	// check if key hot
	if entry.Count >= policy.HotKeys.Threshold {
		go h.Extend(ctx, key, policy.TTL, policy.HotKeys.TTLMultiplier)
	}
}

// TODO: add cleanup logic n stuff

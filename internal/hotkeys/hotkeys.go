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

// rn in v1, it depends on the ttl in the policy config , in extend, which logic can be improved
type HotKeyEntry struct {
	count         int64
	lastIncreased time.Time            // last time the hot key ttl was extended
	hkPolicy      *config.HotKeyPolicy // policy specific fields
	windowEnd     time.Time
}

type hkEvent struct {
	key    string
	policy *config.PolicyConfig
}

// hot key service state
type HotKeyService struct {
	mu              sync.RWMutex
	m               map[string]*HotKeyEntry // map to store key:[count, last increased]
	hkChan          chan *hkEvent           // channel to get events
	maxKeys         int
	redis           redis.Backend // innterface so no ref
	cleanupInterval time.Duration

	// policy specific
	// staleAfter        time.Duration
	// minExtendInterval time.Duration
}

func NewHotKeyService(global *config.GlobalConfig, redisClient redis.Backend, bufSize int) *HotKeyService {
	// if global is disabled then return
	if global.HotKeys == nil || !global.Aegis.HotKeys {
		return nil
	}
	return &HotKeyService{
		m:               make(map[string]*HotKeyEntry),
		hkChan:          make(chan *hkEvent, bufSize),
		maxKeys:         global.HotKeys.MaxTracked,
		redis:           redisClient,
		cleanupInterval: global.HotKeys.CleanupInterval,

		// minExtendInterval: global.HotKeys.MinExtendInterval,
		// staleAfter:        global.HotKeys.StaleAfter,
	}
}

// Start spawns N workers draining the event channel + cleanup goroutine
func (h *HotKeyService) Init(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case ev := <-h.hkChan:
					h.increment(ctx, ev.key, ev.policy)

				}
			}
		}()
	}

	// cleanup goroutine resets counts and evicts stale keys
	go func() {
		ticker := time.NewTicker(h.cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Track enqueues a key event, non-blocking
func (h *HotKeyService) Track(key string, policy *config.PolicyConfig) {
	select {
	case h.hkChan <- &hkEvent{key: key, policy: policy}:
	default: // channel full, drop silently for v1
	}
}

// increment for each hot key
func (h *HotKeyService) increment(ctx context.Context, key string, policy *config.PolicyConfig) {
	if policy == nil || policy.HotKeys == nil || !policy.HotKeys.Enabled {
		return
	}
	now := time.Now()
	h.mu.Lock()

	entry, ok := h.m[key]
	if !ok {
		if len(h.m) >= h.maxKeys {
			// max keys tracked, cant track any more
			h.mu.Unlock()
			return
		}
		// insert new HK entry
		h.m[key] = &HotKeyEntry{
			count:     1,
			windowEnd: now.Add(policy.HotKeys.Window),
			hkPolicy:  policy.HotKeys,
		}
		h.mu.Unlock()
		return
	}

	// window logic
	if entry.windowEnd.IsZero() || now.After(entry.windowEnd) {
		entry.count = 1
		entry.windowEnd = now.Add(entry.hkPolicy.Window)
	} else {
		entry.count++
	}

	// check if hot key is hot
	isHot := entry.count >= entry.hkPolicy.Threshold
	h.mu.Unlock() // release before spawning goroutine

	if isHot {
		h.Extend(ctx, key, policy)
	}
}

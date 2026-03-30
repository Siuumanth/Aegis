package hotkeys

import (
	"context"
	"time"
)

// clean up every internval
func (h *HotKeyService) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, entry := range h.m {
		// check if key is stale
		if time.Since(entry.LastIncreased) > h.staleAfter {
			delete(h.m, key)
			continue
		}
		// reset count so key has to re-earn hot status
		entry.Count = 0
	}
}

// extend ttl logic, ttl * multiplier
func (h *HotKeyService) Extend(ctx context.Context, key string, policyTTL time.Duration, multiplier float64) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry, ok := h.m[key]
	if !ok {
		return nil
	}

	// respect minExtendInterval between extends to avoid hammering redis calls
	if !entry.LastIncreased.IsZero() && time.Since(entry.LastIncreased) < h.minExtendInterval {
		return nil
	}

	newTTL := time.Duration(float64(policyTTL) * multiplier)

	if err := h.redis.Expire(ctx, key, newTTL); err != nil {
		return err
	}

	entry.LastIncreased = time.Now()
	return nil
}

func (h *HotKeyService) Delete(ctx context.Context, key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.m, key)
	h.redis.Del(ctx, key)
}

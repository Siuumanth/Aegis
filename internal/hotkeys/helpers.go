package hotkeys

import (
	"Aegis/config"
	"context"
	"time"
)

// clean up every internval
func (h *HotKeyService) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, entry := range h.m {
		// check if key is stale
		now := time.Now()

		if entry.lastIncreased.IsZero() {
			continue // don’t delete fresh keys
		}

		if now.Sub(entry.lastIncreased) > entry.hkPolicy.StaleAfter {
			delete(h.m, key)
		}
		// no need to reset count as window end handles it
	}
}

// extend ttl logic, ttl * multiplier
// policy specific
// maybe bad design cuz extend knows policy
// TODO: Shud i take max TTL into account?
func (h *HotKeyService) Extend(ctx context.Context, key string, policy *config.PolicyConfig) error {
	if policy.TTL == nil || *policy.TTL == 0 {
		return nil // don’t extend infinite keys
	}
	// check if we need to extend with mutex

	h.mu.Lock()
	entry, ok := h.m[key]

	// If entry is gone or we are within the MinExtendInterval cooldown, stop here
	if !ok || (!entry.lastIncreased.IsZero() && time.Since(entry.lastIncreased) < policy.HotKeys.MinExtendInterval) {
		h.mu.Unlock()
		return nil
	}
	h.mu.Unlock()

	// use proper policy specific fields
	// 2. Prepare the command, no CS
	policyTTL := policy.TTL
	multiplier := policy.HotKeys.TTLMultiplier
	newTTL := time.Duration(float64(*policyTTL) * multiplier)

	// 3. Network call to increase TTL
	if err := h.redis.Expire(ctx, key, newTTL); err != nil {
		return err
	}
	// 4. update the last increased state
	h.mu.Lock()
	// Re-verify the entry still exists, might be cleaned up
	if entry, ok = h.m[key]; ok {
		entry.lastIncreased = time.Now()
	}
	h.mu.Unlock()
	return nil
}

func (h *HotKeyService) Delete(key string) {
	h.mu.Lock()
	delete(h.m, key)
	defer h.mu.Unlock()
}

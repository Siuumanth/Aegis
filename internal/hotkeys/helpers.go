package hotkeys

import (
	"context"
	"time"
)

func (h *HotKeyService) Extend(ctx context.Context, key string, policyTTL time.Duration, multiplier float64) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry, ok := h.m[key]
	if !ok {
		return nil
	}

	// respect min interval between extends
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

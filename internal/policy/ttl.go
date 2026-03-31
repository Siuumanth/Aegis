package policy

import (
	"Aegis/config"
	"time"
)

// resolveTTL determines the final TTL for a SET command.
// prefers client-provided TTL, falls back to policy TTL.
// ResolveTTL determines final TTL
func ResolveTTL(pc *config.PolicyConfig, clientTTL time.Duration) time.Duration {

	// 1. client TTL has highest priority
	if clientTTL > 0 {
		return ClampTTL(pc, clientTTL)
	}

	// 2. policy TTL, n client ttl was 0, use policy TTL
	if pc != nil && pc.TTL != nil {
		// 0 means no expiry → valid
		return ClampTTL(pc, *pc.TTL)
	}

	// 3. fallback means give default ttl
	return config.DefaultTTL
}

// ClampTTL ensures TTL stays within bounds
func ClampTTL(pc *config.PolicyConfig, ttl time.Duration) time.Duration {

	// no expiry → don't clamp
	if ttl == 0 {
		return 0
	}

	if pc != nil && pc.MinTTL != nil && *pc.MinTTL > 0 && ttl < *pc.MinTTL {
		ttl = *pc.MinTTL
	}

	if pc != nil && pc.MaxTTL != nil && *pc.MaxTTL > 0 && ttl > *pc.MaxTTL {
		ttl = *pc.MaxTTL
	}

	return ttl
}

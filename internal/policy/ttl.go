package policy

import (
	"Aegis/config"
	"time"
)

// resolveTTL determines the final TTL for a SET command.
// prefers client-provided TTL, falls back to policy TTL.
func ResolveTTL(pc *config.PolicyConfig, clientTTL time.Duration) time.Duration {
	// client TTL is the ttl in the sent command
	ttl := pc.TTL

	// prefer client TTL if provided
	if clientTTL > 0 {
		ttl = clientTTL
	}

	return ClampTTL(pc, ttl)
}

// ClampTTL ensures TTL stays within policy bounds.
func ClampTTL(pc *config.PolicyConfig, ttl time.Duration) time.Duration {
	if pc.MinTTL > 0 && ttl < pc.MinTTL {
		ttl = pc.MinTTL
	}
	if pc.MaxTTL > 0 && ttl > pc.MaxTTL {
		ttl = pc.MaxTTL
	}
	return ttl
}

package config

import "time"

// ===== TTL Defaults =====
// unbounded
// these ttls are for no expiry
const (
	DefaultTTL    = 60 * time.Second
	DefaultMinTTL = 0 * time.Second
	DefaultMaxTTL = 0 * time.Second
)

// ===== Singleflight =====
const (
	DefaultSingleflightEnabled = false
)

// ===== Hot Key Defaults =====
const (
	DefaultHotKeyEnabled       = false
	DefaultHotKeyWindow        = 2 * time.Second
	DefaultHotKeyThreshold     = 100
	DefaultHotKeyTTLMultiplier = 2.0
)

// ===== System Hot Key Limits =====
const (
	DefaultMaxTrackedKeys    = 10000
	DefaultCleanupInterval   = 10 * time.Second
	DefaultStaleAfter        = 60 * time.Second
	DefaultMinExtendInterval = 5 * time.Second
)

// ===== Redis Client and Server Defaults Defaults =====
const (
	DefaultServerHost         = "0.0.0.0"
	DefaultServerPort         = 6379
	DefaultServerReadTimeout  = 60 * time.Second
	DefaultServerWriteTimeout = 5 * time.Second

	DefaultRedisAddress      = "localhost:6379"
	DefaultRedisPoolSize     = 10
	DefaultRedisMinIdleConns = 2
	DefaultRedisDialTimeout  = 5 * time.Second
	DefaultRedisReadTimeout  = 3 * time.Second
	DefaultRedisWriteTimeout = 3 * time.Second
	DefaultRedisMaxRetries   = 2
)

// WP AND TAG PROCESS defaults
const (
	DefaultTagWorkers    = 4
	DefaultTagBufSize    = 1000
	DefaultHotKeyWorkers = 4
	DefaultHotKeyBufSize = 1000
)

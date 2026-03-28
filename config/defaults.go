package config

import "time"

// ===== TTL Defaults =====
const (
	DefaultTTL    = 60 * time.Second
	DefaultMinTTL = 5 * time.Second
	DefaultMaxTTL = 10 * time.Minute
)

// ===== Singleflight =====
const (
	DefaultSingleflightEnabled = false
)

// ===== Hot Key Defaults =====
const (
	DefaultHotKeyEnabled       = false
	DefaultHotKeyWindow        = 1 * time.Second
	DefaultHotKeyThreshold     = 100
	DefaultHotKeyTTLMultiplier = 2.0
)

// ===== System Hot Key Limits =====
const (
	DefaultMaxTrackedKeys  = 10000
	DefaultCleanupInterval = 10 * time.Second
)

// ===== Redis Client Defaults =====
const (
	DefaultRedisPort         = 6380
	DefaultRedisPoolSize     = 100
	DefaultRedisMinIdleConns = 10
	DefaultRedisMaxRetries   = 2
	DefaultDialTimeout       = 5 * time.Second
	DefaultReadTimeout       = 3 * time.Second
	DefaultWriteTimeout      = 3 * time.Second
)

// ===== Server Defaults =====
const (
	DefaultHost           = "0.0.0.0"
	DefaultPort           = 6379
	DefaultMaxConnections = 10000
	DefaultServerReadTO   = 30 * time.Second
	DefaultServerWriteTO  = 30 * time.Second
)

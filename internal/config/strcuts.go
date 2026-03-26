package config

import "time"

type Config struct {
	Server          ServerConfig          `yaml:"server"`
	Redis           RedisConfig           `yaml:"redis"`
	Resp            RESPConfig            `yaml:"resp"`
	TagInvalidation TagInvalidationConfig `yaml:"tag_invalidation"`
	Singleflight    SingleflightConfig    `yaml:"singleflight"`
	TTLPolicy       TTLPolicyConfig       `yaml:"ttl_policy"`
	SWR             SWRConfig             `yaml:"swr"`
	NegativeCache   NegativeCacheConfig   `yaml:"negative_cache"`
	HotKeys         HotKeysConfig         `yaml:"hot_keys"`
	Retry           RetryConfig           `yaml:"retry"`
	Observability   ObservabilityConfig   `yaml:"observability"`
}

type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	MaxConnections int           `yaml:"max_connections"`
}

type RedisConfig struct {
	Address      string        `yaml:"address"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	MaxRetries   int           `yaml:"max_retries"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type RESPConfig struct {
	Version     int    `yaml:"version"`
	StrictMode  bool   `yaml:"strict_mode"`
	MaxBulkSize string `yaml:"max_bulk_size"`
}

type TagInvalidationConfig struct {
	Enabled           bool `yaml:"enabled"`
	MaxTagsPerKey     int  `yaml:"max_tags_per_key"`
	MaxKeysPerTag     int  `yaml:"max_keys_per_tag"`
	AsyncInvalidation bool `yaml:"async_invalidation"`
	BatchSize         int  `yaml:"batch_size"`
}

type SingleflightConfig struct {
	Enabled           bool          `yaml:"enabled"`
	MaxInflightPerKey int           `yaml:"max_inflight_per_key"`
	Timeout           time.Duration `yaml:"timeout"`
}

type TTLPolicyConfig struct {
	Enabled          bool                     `yaml:"enabled"`
	DefaultTTL       time.Duration            `yaml:"default_ttl"`
	RespectClientTTL bool                     `yaml:"respect_client_ttl"`
	TagTTLs          map[string]time.Duration `yaml:"tag_ttls"`
	KeyPatterns      map[string]time.Duration `yaml:"key_patterns"`
	MinTTL           time.Duration            `yaml:"min_ttl"`
	MaxTTL           time.Duration            `yaml:"max_ttl"`
}

type SWRConfig struct {
	Enabled          bool          `yaml:"enabled"`
	StaleWindow      time.Duration `yaml:"stale_window"`
	AsyncRefresh     bool          `yaml:"async_refresh"`
	RefreshWorkers   int           `yaml:"refresh_workers"`
	RefreshQueueSize int           `yaml:"refresh_queue_size"`
}

type NegativeCacheConfig struct {
	Enabled     bool          `yaml:"enabled"`
	TTL         time.Duration `yaml:"ttl"`
	Marker      string        `yaml:"marker"`
	CacheErrors bool          `yaml:"cache_errors"`
}

type HotKeysConfig struct {
	Enabled         bool          `yaml:"enabled"`
	Window          time.Duration `yaml:"window"`
	Threshold       int           `yaml:"threshold"`
	MaxTrackedKeys  int           `yaml:"max_tracked_keys"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	Actions         HotKeyActions `yaml:"actions"`
}

type HotKeyActions struct {
	ExtendTTL     bool `yaml:"extend_ttl"`
	TTLMultiplier int  `yaml:"ttl_multiplier"`
	SWREnabled    bool `yaml:"swr_enabled"`
}

type RetryConfig struct {
	Enabled     bool          `yaml:"enabled"`
	MaxAttempts int           `yaml:"max_attempts"`
	BaseDelay   time.Duration `yaml:"base_delay"`
	MaxDelay    time.Duration `yaml:"max_delay"`
	Jitter      bool          `yaml:"jitter"`
}

type ObservabilityConfig struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Metrics    MetricsConfig    `yaml:"metrics"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

type MetricsConfig struct {
	CacheHits         bool `yaml:"cache_hits"`
	CacheMisses       bool `yaml:"cache_misses"`
	StaleServed       bool `yaml:"stale_served"`
	Invalidations     bool `yaml:"invalidations"`
	HotKeys           bool `yaml:"hot_keys"`
	SingleflightDedup bool `yaml:"singleflight_dedup"`
	LatencyHistogram  bool `yaml:"latency_histogram"`
}

type LoggingConfig struct {
	Level                string        `yaml:"level"`
	Format               string        `yaml:"format"`
	SlowRequestThreshold time.Duration `yaml:"slow_request_threshold"`
}

package config

import "time"

// top level config
type Config struct {
	Server   ServerConfig  `yaml:"server"`
	Redis    RedisConfig   `yaml:"redis"`
	Defaults DefaultConfig `yaml:"defaults"`
	HotKeys  HotKeysConfig `yaml:"hot_keys"`
	Policies []Policy      `yaml:"policies"`
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
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	MaxRetries   int           `yaml:"max_retries"`
}

// global defaults applied to all matched keys unless overridden

type DefaultConfig struct {
	TTL          time.Duration `yaml:"ttl"`
	MinTTL       time.Duration `yaml:"min_ttl"`
	MaxTTL       time.Duration `yaml:"max_ttl"`
	Singleflight bool          `yaml:"singleflight"`
}

// system-wide hot key settings, per-pattern config lives in Policy
type HotKeysConfig struct {
	MaxTracked      int           `yaml:"max_tracked"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

// one policy block in the policies list
type Policy struct {
	Name   string       `yaml:"name"`
	Match  MatchConfig  `yaml:"match"`
	Config PolicyConfig `yaml:"config"`
}

type MatchConfig struct {
	Pattern string `yaml:"pattern"` // glob, e.g. "user:*"
	Tag     string `yaml:"tag"`     // tag name, e.g. "users"
}

type PolicyConfig struct {
	TTL          time.Duration `yaml:"ttl"`
	MinTTL       time.Duration `yaml:"min_ttl"`
	MaxTTL       time.Duration `yaml:"max_ttl"`
	Singleflight bool          `yaml:"singleflight"`
	Tags         []string      `yaml:"tags"`
	HotKeys      HotKeyPolicy  `yaml:"hot_key"`
}

type HotKeyPolicy struct {
	Enabled           bool          `yaml:"enabled"`
	Window            time.Duration `yaml:"window"`
	Threshold         int64         `yaml:"threshold"`
	TTLMultiplier     float64       `yaml:"ttl_multiplier"`
	minExtendInterval time.Duration `yaml:"min_extend_interval"`
}

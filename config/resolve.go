package config

import (
	"time"
)

// TODO: update defaults and resolving
// TODO: Default stale after is multiplier * TTL
// for default and hot keys
type GlobalConfig struct {
	HotKeys  HotKeysConfig
	Defaults DefaultConfig
}

// the final config which will be passed to all functions
type RuntimeConfig struct {
	GlobalConfig    GlobalConfig
	PatternPolicies map[string]PolicyConfig
	TagPolicies     map[string]PolicyConfig
}

// BuildRuntimeConfig converts raw YAML config → runtime maps
func BuildRuntimeConfig(cfg *Config) *RuntimeConfig {
	rt := &RuntimeConfig{
		GlobalConfig:    GlobalConfig{HotKeys: cfg.HotKeys, Defaults: cfg.Defaults},
		PatternPolicies: make(map[string]PolicyConfig),
		TagPolicies:     make(map[string]PolicyConfig),
	}

	// set default of global
	mergeGlobal(&rt.GlobalConfig)

	for _, p := range cfg.Policies {
		mergeDefaults(cfg, &p.Config)

		pc := p.Config

		// normalize TTL bounds
		if pc.MinTTL > 0 && pc.TTL < pc.MinTTL {
			pc.TTL = pc.MinTTL
		}
		if pc.MaxTTL > 0 && pc.TTL > pc.MaxTTL {
			pc.TTL = pc.MaxTTL
		}

		// pattern-based
		if p.Match.Pattern != "" {
			rt.PatternPolicies[p.Match.Pattern] = pc
		}

		// tag-based
		if p.Match.Tag != "" {
			rt.TagPolicies[p.Match.Tag] = pc
		}
	}

	return rt
}

// merge defaults into policy config
func mergeDefaults(cfg *Config, pc *PolicyConfig) {

	// prefer explicit, else fallback
	if !pc.Singleflight {
		// auto false if not present
		pc.Singleflight = cfg.Defaults.Singleflight
	}

	// if hot key is enabled then chesck and use defaults
	if pc.HotKeys.Enabled {
		if pc.HotKeys.Window == 0 {
			pc.HotKeys.Window = DefaultHotKeyWindow
		}
		if pc.HotKeys.Threshold == 0 {
			pc.HotKeys.Threshold = DefaultHotKeyThreshold
		}
		if pc.HotKeys.TTLMultiplier == 0 {
			pc.HotKeys.TTLMultiplier = DefaultHotKeyTTLMultiplier
		}
	}

	// do same for all other values like ttl, min_ttl, max_ttl...
	// but if ttl is not defined, it will be 0 automatically and will be ignored
	// but lets make a default custom logic
	// configute ttls
	pc.TTL = pickDuration(pc.TTL, pickDuration(cfg.Defaults.TTL, DefaultTTL))
	pc.MinTTL = pickDuration(pc.MinTTL, pickDuration(cfg.Defaults.MinTTL, DefaultMinTTL))
	pc.MaxTTL = pickDuration(pc.MaxTTL, pickDuration(cfg.Defaults.MaxTTL, DefaultMaxTTL))

}

func mergeGlobal(global *GlobalConfig) {
	// check hot keys
	if global.HotKeys.MaxTracked == 0 {
		global.HotKeys.MaxTracked = DefaultMaxTrackedKeys
	}
	if global.HotKeys.CleanupInterval == 0 {
		global.HotKeys.CleanupInterval = DefaultCleanupInterval
	}
}

func pickDuration(a, b time.Duration) time.Duration {
	if a != 0 {
		return a
	}
	return b
}

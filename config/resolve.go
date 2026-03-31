package config

import (
	"time"
)

// for default and hot keys
type GlobalConfig struct {
	HotKeys  *HotKeysConfig
	Defaults *DefaultConfig
}

// the final config which will be passed to all functions
type RuntimeConfig struct {
	GlobalConfig    *GlobalConfig
	PatternPolicies map[string]PolicyConfig
}

// BuildRuntimeConfig converts raw YAML config → runtime maps
func BuildRuntimeConfig(cfg *Config) *RuntimeConfig {
	if cfg.Aegis == nil {
		cfg.Aegis = &Aegis{}
	}
	// first step, check if features are true
	if !cfg.Aegis.HotKeys {
		cfg.HotKeys = nil
	}
	rt := &RuntimeConfig{
		GlobalConfig:    &GlobalConfig{HotKeys: cfg.HotKeys, Defaults: cfg.Defaults},
		PatternPolicies: make(map[string]PolicyConfig),
	}

	// set default of global
	mergeGlobal(rt.GlobalConfig)

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
		// boool defaults to false so thats good
		// check aegis features enabled or not and make it nil
		if !cfg.Aegis.HotKeys {
			pc.HotKeys = nil
		}
		if !cfg.Aegis.Tags {
			pc.Tags = nil
		}
		if !cfg.Aegis.Singleflight {
			pc.Singleflight = false
		}

		// pattern-based
		if p.Match.Pattern != "" {
			rt.PatternPolicies[p.Match.Pattern] = pc
		}

	}

	return rt
}

// merge defaults into policy config
func mergeDefaults(cfg *Config, pc *PolicyConfig) {

	// prefer explicit, else fallback
	if !pc.Singleflight {
		// auto false if not present
		pc.Singleflight = DefaultSingleflightEnabled
	}

	// if hot key is enabled then chesck and use defaults
	if pc.HotKeys != nil && pc.HotKeys.Enabled {
		if pc.HotKeys.Window == 0 {
			pc.HotKeys.Window = DefaultHotKeyWindow
		}
		if pc.HotKeys.Threshold == 0 {
			pc.HotKeys.Threshold = DefaultHotKeyThreshold
		}
		if pc.HotKeys.TTLMultiplier == 0 {
			pc.HotKeys.TTLMultiplier = DefaultHotKeyTTLMultiplier
		}
		// min extend interval and stale after
		if pc.HotKeys.MinExtendInterval == 0 && cfg.HotKeys != nil {
			// give the global value if not assigned
			pc.HotKeys.MinExtendInterval = cfg.HotKeys.MinExtendInterval
		}
		if pc.HotKeys.StaleAfter == 0 {
			// default to multiplier * TTL
			pc.HotKeys.StaleAfter = pc.TTL * time.Duration(pc.HotKeys.TTLMultiplier)
		}
	}

	// do same for all other values like ttl, min_ttl, max_ttl...
	// but if ttl is not defined, it will be 0 automatically and will be ignored
	// but lets make a default custom logic
	// configute ttls
	if cfg.Defaults == nil {
		return
	}
	pc.TTL = pickDuration(pc.TTL, pickDuration(cfg.Defaults.TTL, DefaultTTL))
	pc.MinTTL = pickDuration(pc.MinTTL, pickDuration(cfg.Defaults.MinTTL, DefaultMinTTL))
	pc.MaxTTL = pickDuration(pc.MaxTTL, pickDuration(cfg.Defaults.MaxTTL, DefaultMaxTTL))

}

func mergeGlobal(global *GlobalConfig) {
	if global.HotKeys == nil {
		return
	}
	// check hot keys
	if global.HotKeys.MaxTracked == 0 {
		global.HotKeys.MaxTracked = DefaultMaxTrackedKeys
	}
	if global.HotKeys.CleanupInterval == 0 {
		global.HotKeys.CleanupInterval = DefaultCleanupInterval
	}
	if global.HotKeys.StaleAfter == 0 {
		// default to multiplier * TTL
		global.HotKeys.StaleAfter = DefaultStaleAfter
	}
	if global.HotKeys.MinExtendInterval == 0 {
		global.HotKeys.MinExtendInterval = DefaultMinExtendInterval
	}
}

func pickDuration(a, b time.Duration) time.Duration {
	if a != 0 {
		return a
	}
	return b
}

package config

import (
	"time"
)

/*
BEHAVIOUR:
TTL:
nil  → fallback (defaults → system)
0    → explicit no expiry
>0   → use value

MinTTL / MaxTTL:
nil  → fallback
>0   → apply

HotKeys:
uses resolved TTL correctly
*/

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
	// redundant ig
	if cfg.Aegis == nil {
		cfg.Aegis = &Aegis{}
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

		// normalize TTL bounds (only if ttl > 0)
		if pc.TTL != nil && *pc.TTL > 0 {
			if pc.MinTTL != nil && *pc.MinTTL > 0 && *pc.TTL < *pc.MinTTL {
				val := *pc.MinTTL
				pc.TTL = &val
			}
			if pc.MaxTTL != nil && *pc.MaxTTL > 0 && *pc.TTL > *pc.MaxTTL {
				val := *pc.MaxTTL
				pc.TTL = &val
			}
		}

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
// merge defaults into policy config
func mergeDefaults(cfg *Config, pc *PolicyConfig) {

	// prefer explicit, else fallback
	if !pc.Singleflight {
		pc.Singleflight = cfg.Aegis.Singleflight // its false by default
	}

	// ---------------- TTL RESOLUTION ----------------

	if pc.TTL == nil {
		// fallback chain
		if cfg.Defaults != nil && cfg.Defaults.TTL != nil {
			val := *cfg.Defaults.TTL
			pc.TTL = &val
		} else {
			val := DefaultTTL
			pc.TTL = &val
		}
	}
	// pc.TTL == 0 explicitly means noo expiry, DO NOT override

	// ---------------- MIN TTL ----------------

	if pc.MinTTL == nil {
		if cfg.Defaults != nil && cfg.Defaults.MinTTL != nil {
			val := *cfg.Defaults.TTL
			pc.TTL = &val
		} else {
			val := DefaultMinTTL
			pc.MinTTL = &val
		}
	}

	// ---------------- MAX TTL ----------------

	if pc.MaxTTL == nil {
		if cfg.Defaults != nil && cfg.Defaults.MaxTTL != nil {
			val := *cfg.Defaults.TTL
			pc.TTL = &val
		} else {
			val := DefaultMaxTTL
			pc.MaxTTL = &val
		}
	}

	// if hot key is enabled then check and use defaults
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

		// min extend interval
		if pc.HotKeys.MinExtendInterval == 0 && cfg.HotKeys != nil {
			pc.HotKeys.MinExtendInterval = cfg.HotKeys.MinExtendInterval
		}

		// stale after
		if pc.HotKeys.StaleAfter == 0 {
			var ttl time.Duration
			if pc.TTL != nil {
				ttl = *pc.TTL
			}
			pc.HotKeys.StaleAfter = time.Duration(float64(ttl) * pc.HotKeys.TTLMultiplier)
		}
	}
}

func mergeGlobal(global *GlobalConfig) {
	if global.HotKeys == nil {
		return
	}

	if global.HotKeys.MaxTracked == 0 {
		global.HotKeys.MaxTracked = DefaultMaxTrackedKeys
	}
	if global.HotKeys.CleanupInterval == 0 {
		global.HotKeys.CleanupInterval = DefaultCleanupInterval
	}
	if global.HotKeys.StaleAfter == 0 {
		global.HotKeys.StaleAfter = DefaultStaleAfter
	}
	if global.HotKeys.MinExtendInterval == 0 {
		global.HotKeys.MinExtendInterval = DefaultMinExtendInterval
	}
}

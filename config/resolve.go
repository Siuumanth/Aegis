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
	Aegis    *Aegis
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
	// TODO: ORganise code better
	if cfg.Server == nil {
		cfg.Server = &ServerConfig{}
	}
	if cfg.Redis == nil {
		cfg.Redis = &RedisConfig{}
	}

	// apply redis and server defaults
	applyServerDefaults(cfg.Server)
	applyRedisDefaults(cfg.Redis)
	if cfg.Aegis == nil {
		cfg.Aegis = &Aegis{}
	}

	if cfg.Aegis.HotKeys == false {
		cfg.HotKeys = nil
	} else {
		if cfg.HotKeys == nil {
			cfg.HotKeys = &HotKeysConfig{}
		}
	}

	rt := &RuntimeConfig{
		GlobalConfig:    &GlobalConfig{HotKeys: cfg.HotKeys, Defaults: cfg.Defaults, Aegis: cfg.Aegis},
		PatternPolicies: make(map[string]PolicyConfig),
	}

	// set default of global
	mergeGlobal(rt.GlobalConfig)

	for _, p := range cfg.Policies {
		// if default hot key is true then set hot key policy as
		// enabled
		if cfg.Aegis != nil && cfg.Aegis.HotKeys == true {
			//fmt.Println("T1")
			if p.Config.HotKeys == nil || (p.Config.HotKeys != nil &&
				!p.Config.HotKeys.Enabled) {
				//	fmt.Println("T2S for ", p.Config)
				// init a blank policy
				p.Config.HotKeys = &HotKeyPolicy{Enabled: true}
			}
		}
		mergeDefaults(cfg, &p.Config)
		pc := p.Config
		// if hot keys is enabled, create a default policy hk
		if cfg.Aegis.HotKeys && pc.HotKeys == nil {
			pc.HotKeys = &HotKeyPolicy{}
		}

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

		// final cleanup, kept at last for easier understanding
		// check aegis features enabled or not and make it nil
		if !cfg.Aegis.HotKeys || (pc.HotKeys != nil && !pc.HotKeys.Enabled) {
			pc.HotKeys = nil
		}
		if !cfg.Aegis.Tags || pc.Tags == nil {
			pc.Tags = nil
		}
		if !cfg.Aegis.Singleflight {
			pc.Singleflight = DefaultSingleflightEnabled
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
	// per policy config

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
	if cfg.Aegis.HotKeys && pc.HotKeys != nil && pc.HotKeys.Enabled {

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
	if global.HotKeys == nil || global.Aegis.HotKeys == false {
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
	if global.HotKeys.Window == 0 {
		global.HotKeys.Window = DefaultHotKeyWindow
	}
	if global.HotKeys.Threshold == 0 {
		global.HotKeys.Threshold = DefaultHotKeyThreshold
	}
	if global.HotKeys.TTLMultiplier == 0 {
		global.HotKeys.TTLMultiplier = DefaultHotKeyTTLMultiplier
	}

}

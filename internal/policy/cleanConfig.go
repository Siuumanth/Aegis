package policy

import (
	"Aegis/internal/shared"
	"time"
)

// the final config which will be passed to all functions
type RuntimeConfig struct {
	Global          GlobalConfig
	PatternPolicies map[string]PolicyConfig
	TagPolicies     map[string]PolicyConfig
}

// BuildRuntimeConfig converts raw YAML config → runtime maps
func BuildRuntimeConfig(cfg *Config) *RuntimeConfig {
	rt := &RuntimeConfig{
		PatternPolicies: make(map[string]PolicyConfig),
		TagPolicies:     make(map[string]PolicyConfig),
	}

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
	if pc.TTL == 0 {
		pc.TTL = cfg.Defaults.TTL
	}

	// prefer explicit, else fallback
	if !pc.Singleflight {
		// auto false if not present
		pc.Singleflight = cfg.Defaults.Singleflight
	}

	// if hot key is enabled then check and use defaults
	if pc.HotKey.Enabled {
		if pc.HotKey.Window == 0 {
			pc.HotKey.Window = shared.DefaultHotKeyWindow
		}
		if pc.HotKey.Threshold == 0 {
			pc.HotKey.Threshold = shared.DefaultHotKeyThreshold
		}
		if pc.HotKey.TTLMultiplier == 0 {
			pc.HotKey.TTLMultiplier = shared.DefaultHotKeyTTLMultiplier
		}
	}

	// do same for all other values like ttl, min_ttl, max_ttl...
	// but if ttl is not defined, it will be 0 automatically and will be ignored
	// but lets make a default custom logic
	// configute ttls
	pc.TTL = pickDuration(pc.TTL, pickDuration(cfg.Defaults.TTL, shared.DefaultTTL))
	pc.MinTTL = pickDuration(pc.MinTTL, pickDuration(cfg.Defaults.MinTTL, shared.DefaultMinTTL))
	pc.MaxTTL = pickDuration(pc.MaxTTL, pickDuration(cfg.Defaults.MaxTTL, shared.DefaultMaxTTL))

}

func pickDuration(a, b time.Duration) time.Duration {
	if a != 0 {
		return a
	}
	return b
}

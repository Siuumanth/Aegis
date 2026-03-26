package policy

// the final config which will be passed to all functions
type RuntimeConfig struct {
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
		pc := mergeDefaults(cfg.Defaults, p.Config)

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
func mergeDefaults(def DefaultConfig, pc PolicyConfig) PolicyConfig {
	if pc.TTL == 0 {
		pc.TTL = def.TTL
	}

	// prefer explicit, else fallback
	if !pc.Singleflight {
		pc.Singleflight = def.Singleflight
	}

	if !pc.HotKey.Enabled {
		pc.HotKey = defaultHotKey // if you define one
	}

	return pc
}

package policy

import (
	"Aegis/config"
	"Aegis/internal/resp"
	"path"
)

// this respnsible for matching the pattern
// Takes in command , config and returns the populated request

/*
type RuntimeConfig struct {
	GlobalConfig    GlobalConfig
	PatternPolicies map[string]PolicyConfig
	TagPolicies     map[string]PolicyConfig
}
*/

type Engine struct {
	cfg *config.RuntimeConfig
}

func NewEngine(cfg *config.RuntimeConfig) *Engine {
	return &Engine{cfg: cfg}
}

func (e *Engine) Match(cmd *resp.Command) *config.PolicyConfig {
	// match policy to pattern, user:*
	// loop over cfgs policies
	// handles cases like users:*:profiles
	for pattern, policy := range e.cfg.PatternPolicies {
		matched, err := path.Match(pattern, cmd.Key)
		if err != nil {
			// invalid pattern, skip
			continue
		}
		if matched {
			// return copy, not pointer to map value
			pc := policy
			return &pc
		}
	}
	return nil
}

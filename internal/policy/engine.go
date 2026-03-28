package policy

import (
	"Aegis/config"
	"Aegis/internal/resp"
	"strings"
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

	for pattern, policy := range e.cfg.PatternPolicies {
		prefix := strings.TrimSuffix(pattern, "*")

		if strings.HasPrefix(cmd.Key, prefix) {
			return &policy
		}
	}

	return nil // no matches, blank

}

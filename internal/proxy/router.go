package proxy

import (
	"Aegis/config"
	"Aegis/internal/handler"
	"Aegis/internal/policy"
	"Aegis/internal/resp"
	"Aegis/internal/types"
	"net"
	"strings"
)

// Dependencies ≠ per-request data
// struct shud hold depenedices
type Router struct {
	cfg     *config.RuntimeConfig
	policy  *policy.Engine
	handler *handler.Handler
}

// the router matches pattern and then routes it
func NewRouter(cfg *config.RuntimeConfig) *Router {
	engine := policy.NewEngine(cfg)
	return &Router{
		cfg:    cfg,
		policy: engine,
	}
}

// match policy, route based on cmd
func (r *Router) Route(cmd *resp.Command, conn net.Conn) error {

	// 1. match policy
	match := r.policy.Match(cmd)

	req := &types.Request{
		Cmd:    cmd,
		Policy: match,
		Conn:   conn,
	}
	// request is all  the downstream processes need to know to match
	// 2. routing decision happens in handler layer (not here)

	switch strings.ToUpper(cmd.Name) {
	case "GET":
		return r.handler.Get(req)
	case "SET":
		return r.handler.Set(req)
		// case "DEL":
		// 	return r.handler.Del(req)
		// case "AEGIS.INVALIDATE":
		// 	return r.handler.Invalidate(req)
		// default:
		// 	return r.handler.Passthrough(req)
	}
	return nil
}

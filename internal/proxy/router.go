package proxy

import (
	"Aegis/config"
	"Aegis/internal/policy"
	"Aegis/internal/resp"
	"net"
)

type Request struct {
	Cmd    *resp.Command
	Policy *config.PolicyConfig // nil if no match
	Conn   net.Conn
}

type Router struct {
	cfg    *config.RuntimeConfig
	policy *policy.Engine
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
func (r *Router) Route(cmd *resp.Command, conn net.Conn) *Request {

	// 1. match policy
	match := r.policy.Match(cmd)

	req := &Request{
		Cmd:    cmd,
		Policy: match,
		Conn:   conn,
	}
	// request is all  the downstream processes need to know to match
	// 2. routing decision happens in handler layer (not here)

	return req
}

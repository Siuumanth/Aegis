package proxy

import (
	"Aegis/config"
	"Aegis/internal/handler"
	"Aegis/internal/policy"
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
	"io"
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
func NewRouter(cfg *config.RuntimeConfig, h *handler.Handler, p *policy.Engine) *Router {
	return &Router{
		cfg:     cfg,
		policy:  p,
		handler: h,
	}
}

// match policy, route based on cmd
func (r *Router) Route(ctx context.Context, cmd *resp.Command, w io.Writer) error {

	// 1. match policy
	match := r.policy.Match(cmd)

	req := &handler.Request{
		Cmd:    cmd,
		Policy: match,
		Writer: w,
	}
	// request is all  the downstream processes need to know to match
	// 2. routing decision based on cmd
	CMD := strings.ToUpper(cmd.Name)
	switch CMD {
	case "GET":
		return r.handler.Get(ctx, req)
	case "SET":
		return r.handler.Set(ctx, req)
	case "HELLO":
		resp.WriteAny(w, []any{"server", "redis"})
		return nil
	case "CLIENT":
		resp.WriteOK(w)
		return nil
	case "DEL":
		return r.handler.Del(ctx, req)

		// not wokring properly
		// BIG TODO: find some solution for bidirectional stuff like pubsub
	// case "SUBSCRIBE", "PSUBSCRIBE":
	// 	return r.handler.PubSubTunnel(ctx, req)

	default:
		if strings.HasPrefix(strings.ToUpper(CMD), "AEGIS.") {
			return r.RouteCustom(ctx, CMD, req)
		}
		err := r.handler.DefaultHandler(ctx, req)
		if err != nil {
			return resp.WriteError(w, shared.ErrBackend)
		}
	}
	return nil
}

func (r *Router) RouteCustom(ctx context.Context, cmd string, req *handler.Request) error {
	switch strings.ToUpper(cmd) {
	case "AEGIS.INVALIDATE":
		return r.handler.Invalidate(ctx, req)
	default:
		return resp.WriteError(req.Writer, shared.ErrInvalidCommand)
	}
}

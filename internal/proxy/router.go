package proxy

import (
	"Aegis/config"
	"Aegis/internal/handler"
	"Aegis/internal/policy"
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
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
func NewRouter(cfg *config.RuntimeConfig, h *handler.Handler, p *policy.Engine) *Router {
	return &Router{
		cfg:     cfg,
		policy:  p,
		handler: h,
	}
}

// match policy, route based on cmd
func (r *Router) Route(ctx context.Context, cmd *resp.Command, conn net.Conn) error {

	// 1. match policy
	match := r.policy.Match(cmd)

	req := &handler.Request{
		Cmd:    cmd,
		Policy: match,
		Conn:   conn,
	}
	// request is all  the downstream processes need to know to match
	// 2. routing decision based on cmd
	CMD := strings.ToUpper(cmd.Name)
	switch CMD {
	case "GET":
		return r.handler.Get(req)
	case "SET":
		return r.handler.Set(req)
	case "HELLO":
		conn.Write([]byte("*2\r\n$6\r\nserver\r\n$5\r\nredis\r\n"))
		return nil
	case "CLIENT":
		conn.Write([]byte("+OK\r\n"))
		return nil
	case "DEL":
		return r.handler.Del(req)

	default:
		if strings.HasPrefix(CMD, "AEGIS.") {
			return r.RouteCustom(CMD, req)
		}
		result, err := r.handler.redis.PassThrough(ctx, cmd)
		if err != nil {
			return resp.WriteError(conn, shared.ErrBackend)
		}
		return resp.WriteAny(conn, result)
	}
}

func (r *Router) RouteCustom(cmd string, req *handler.Request) error {
	switch cmd {
	case "AEGIS.INVALIDATE":
		return r.handler.Invalidate(req)
	default:
		return resp.WriteError(req.Conn, shared.ErrInvalidCommand)
	}
}

package handler

import (
	"Aegis/internal/policy"
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
	"strconv"
	"strings"
	"time"
)

/*
SET:

	policy.ResolveTTL(req.Policy, cmd.ClientTTL)
	→ redis.Set(key, value, ttl)
	→ tags.Register(key, policy.Tags, cmd.ATags)
	→ resp.Write(OK)
*/
func (h *Handler) Set(ctx context.Context, req *Request) error {
	// if no ttl in yaml then we gotta do no ttl modificiation
	// parse client TTL from args if provided (EX 300)
	var ttl time.Duration
	clientTTL := parseClientTTL(req.Cmd.Args)
	if req.Policy != nil && *req.Policy.TTL != 0 {
		// resolve final TTL against policy bounds
		ttl = policy.ResolveTTL(req.Policy, clientTTL)
	} else {
		ttl = clientTTL
	}

	value := req.Cmd.Args[0] // value being set
	if err := h.redis.Set(ctx, req.Cmd.Key, value, ttl); err != nil {
		return resp.WriteError(req.Conn, shared.ErrBackend)
	}
	args := req.Cmd.Args[1:] // passing only modifiers
	// register tags async, nil safe
	if h.tags != nil && req.Policy != nil && req.Policy.Tags != nil {
		h.tags.Register(req.Cmd.Key, req.Policy.Tags, args)
	}

	return resp.WriteOK(req.Conn)
}

// tihs func  scans args for EX or PX and returns th time
func parseClientTTL(args []string) time.Duration {
	for i := 0; i < len(args)-1; i++ {
		switch strings.ToUpper(args[i]) {
		case "EX":
			if secs, err := strconv.Atoi(args[i+1]); err == nil {
				return time.Duration(secs) * time.Second
			}
		case "PX":
			if ms, err := strconv.Atoi(args[i+1]); err == nil {
				return time.Duration(ms) * time.Millisecond
			}
		}
	}
	return 0
}

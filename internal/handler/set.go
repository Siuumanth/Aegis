package handler

import (
	"Aegis/internal/policy"
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

func (h *Handler) Set(req *Request) error {
	// parse client TTL from args if provided (EX 300)
	clientTTL := parseClientTTL(req.Cmd.Args)

	// resolve final TTL against policy bounds
	ttl := policy.ResolveTTL(req.Policy, clientTTL)

	// TODO
	value := req.Cmd.Args[0]
	err := h.redis.Set(context.TODO(), req.Cmd.Key, value, ttl)
	if err != nil {
		return err
	}
	// send Redis-style OK
	req.Conn.Write([]byte("+OK\r\n"))

	return nil
}

// parseClientTTL scans args for EX or PX
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

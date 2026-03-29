package handler

import (
	"Aegis/internal/types"
	"context"
	"fmt"
)

// define get handler:
/*
GET:
    singleflight.Do(key, redis.Get)
    → redis.Get(key)
    → resp.Write(result)
    → hotkeys.Track(key)        ← async, worker pool
*/

func (h *Handler) Get(req *types.Request) error {
	val, err := h.redis.Get(context.TODO(), req.Cmd.Key)
	if err != nil {
		req.Conn.Write([]byte("$-1\r\n")) // nil
		return nil
	}

	// RESP bulk string
	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
	req.Conn.Write([]byte(resp))

	return err
}

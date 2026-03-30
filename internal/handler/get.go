package handler

import (
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
)

// define get handler:
/*
GET:
    singleflight.Do(key, redis.Get)
    → redis.Get(key)
    → resp.Write(result)
    → hotkeys.Track(key)        ← async, worker pool
*/

func (h *Handler) Get(req *Request) error {
	// 1. Send singleflight
	var (
		val string
		err error
	)
	ctx := context.TODO()
	if h.sf != nil {
		var result any
		result, err = h.sf.Do(ctx, req.Cmd.Key, func() (any, error) {
			return h.redis.Get(ctx, req.Cmd.Key)
		})
		if err == nil {
			val = result.(string)
		}
	} else {
		val, err = h.redis.Get(ctx, req.Cmd.Key)
	}

	if err != nil {
		if err == shared.ErrGoRedisNil {
			return resp.WriteNull(req.Conn)
		}
		return resp.WriteError(req.Conn, shared.ErrBackend)
	}

	// track hot key async, nil safe
	if h.hotkeys != nil {
		h.hotkeys.Track(req.Cmd.Key, req.Policy)
	}

	// 2. Send response
	return resp.WriteString(req.Conn, val)
}

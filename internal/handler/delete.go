package handler

import (
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
)

func (h *Handler) Del(ctx context.Context, req *Request) error {

	// 1. delete key from Redis
	if err := h.redis.Del(ctx, req.Cmd.Key); err != nil {
		return resp.WriteError(req.Conn, shared.ErrBackend)
	}

	// 2. async tag cleanup
	if h.tags != nil {
		h.tags.Delete(req.Cmd.Key)
	}

	// 3.  hotkey cleanup
	if h.hotkeys != nil {
		h.hotkeys.Delete(ctx, req.Cmd.Key)
	}

	// 4. RESP OK
	return resp.WriteOK(req.Conn)
}

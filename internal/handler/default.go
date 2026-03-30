package handler

import (
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
)

func (h *Handler) DefaultHandler(req *Request) error {
	ctx := context.TODO()
	result, err := h.redis.PassThrough(ctx, req.Cmd)
	if err != nil {
		return resp.WriteError(req.Conn, shared.ErrBackend)
	}
	// return result back
	return resp.WriteAny(req.Conn, result)
}

package handler

import (
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
)

func (h *Handler) Invalidate(req *Request) error {
	ctx := context.TODO()

	// validate args
	if len(req.Cmd.Args) < 1 {
		return resp.WriteError(req.Conn, shared.ErrInvalidCommand)
	}

	tag := req.Cmd.Key // key is the tag name

	// call tag service
	count, err := h.tags.Invalidate(ctx, tag)
	if err != nil {
		return resp.WriteError(req.Conn, shared.ErrBackend)
	}

	// return number of keys deleted (RESP integer)
	return resp.WriteInteger(req.Conn, count)
}

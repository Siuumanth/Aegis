package handler

import (
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
	"fmt"
)

func (h *Handler) Invalidate(ctx context.Context, req *Request) error {
	// 1. gather all tags to invalidate
	// The first tag is in req.Cmd.Key, the rest are in req.Cmd.Args
	tags := make([]string, 0, len(req.Cmd.Args)+1)
	if req.Cmd.Key != "" {
		tags = append(tags, req.Cmd.Key)
	}
	tags = append(tags, req.Cmd.Args...)

	if len(tags) == 0 {
		return resp.WriteError(req.Writer, shared.ErrInvalidCommand)
	}

	var totalDeleted int64

	// 2. loop and Invalidate
	for _, tag := range tags {
		count, err := h.tags.Invalidate(ctx, tag)
		if err != nil {
			// Log the error but keep going for other tags?
			// Or return immediately? Usually, in a proxy, we log and continue.
			fmt.Printf("Error invalidating tag %s: %v\n", tag, err)
			continue
		}
		totalDeleted += count
	}

	// 3. Return the total sum of keys deleted across all tags
	return resp.WriteInteger(req.Writer, totalDeleted)
}

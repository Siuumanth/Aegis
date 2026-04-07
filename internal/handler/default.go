package handler

import (
	"Aegis/internal/resp"
	"Aegis/internal/shared"
	"context"
	"fmt"
	"io"
	"net"
)

func (h *Handler) DefaultHandler(ctx context.Context, req *Request) error {
	result, err := h.redis.PassThrough(ctx, req.Cmd)
	if err != nil {
		return resp.WriteError(req.Conn, shared.ErrBackend)
	}
	// return result back
	return resp.WriteAny(req.Conn, result)
}

// temperory fix to enable pubsub
func (h *Handler) PubSubTunnel(ctx context.Context, req *Request) error {
	backendConn, err := net.Dial("tcp", h.redisAddr)
	if err != nil {
		return err
	}

	// send the original SUBSCRIBE command to Redis
	n, err := backendConn.Write(req.Cmd.Raw) // raw RESP bytes

	if err != nil {
		return err
	}
	fmt.Println("Raw cmd is ", req.Cmd.Raw, "written", n, "bytes")

	// client to redis
	go func() {
		io.Copy(backendConn, req.Conn)
		backendConn.Close() // close only this side
	}()

	// redis to client
	io.Copy(req.Conn, backendConn)
	req.Conn.Close() // close after stream ends

	return nil
}

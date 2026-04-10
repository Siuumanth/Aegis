package handler

import (
	"Aegis/internal/resp"
	"context"

	"github.com/redis/go-redis/v9"
)

func (h *Handler) DefaultHandler(ctx context.Context, req *Request) error {
	result, err := h.redis.PassThrough(ctx, req.Cmd)
	if err == redis.Nil {
		return resp.WriteNull(req.Writer) // $-1
	}
	if err != nil {
		return resp.WriteError(req.Writer, err)
	}
	// return result back
	return resp.WriteAny(req.Writer, result)
}

// temperory fix to enable pubsub
// func (h *Handler) PubSubTunnel(ctx context.Context, req *Request) error {
// 	backendConn, err := net.Dial("tcp", h.redisAddr)
// 	if err != nil {
// 		return err
// 	}

// 	// send the original SUBSCRIBE command to Redis
// 	n, err := backendConn.Write(req.Cmd.Raw) // raw RESP bytes

// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Raw cmd is ", req.Cmd.Raw, "written", n, "bytes")

// 	// client to redis
// 	go func() {
// 		io.Copy(backendConn, req.Writer)
// 		backendConn.Close() // close only this side
// 	}()

// 	// redis to client
// 	io.Copy(req.Writer, backendConn)
// 	req.Writer.Close() // close after stream ends

// 	return nil
// }

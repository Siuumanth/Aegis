package proxy

import (
	"Aegis/internal/resp"
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// Conn handles a single client connection lifecycle
type Conn struct {
	conn   net.Conn
	parser *resp.Parser
	router *Router
}

func NewConn(conn net.Conn, router *Router, parser *resp.Parser) *Conn {
	return &Conn{
		conn:   conn,
		parser: parser,
		router: router,
	}
}

// Handle reads commands in a loop until client disconnects
func (c *Conn) Handle(globalCtx context.Context) {
	ctx, cancel := context.WithCancel(globalCtx)
	defer cancel()
	defer c.conn.Close()

	// Each iteration = one Redis command
	for {
		// 1. Respect shutdown
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 2. Set read deadline (prevents dead connections)
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// 3. Parse request
		cmd, err := c.parser.Parse()
		if err != nil {
			// client disconnected or read error, exit loop
			//if the client actually disconnects, the Read function will immediately stop waiting and return a specific error (io.EOF)
			if err == io.EOF {
				// normal disconnect
				return
			}

			// unexpected error
			fmt.Printf("parse error: %v\n", err)
			return
		}

		// 4. Route request
		if err := c.router.Route(ctx, cmd, c.conn); err != nil {
			// if write fails → client likely gone
			fmt.Printf("routing error: %v\n", err)

			// optional: break instead of continue
			return
		}
	}
}

/*
-> more notes in tcpNreader
*/

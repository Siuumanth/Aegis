package proxy

import (
	"Aegis/internal/resp"
	"context"
	"net"
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
func (c *Conn) Handle() {
	// context tied to connection lifetime
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer c.conn.Close()

	for {
		cmd, err := c.parser.Parse()
		if err != nil {
			// client disconnected or read error, exit loop
			return
		}

		if err := c.router.Route(ctx, cmd, c.conn); err != nil {
			// log error, keep going unless fatal
			continue
		}
	}
}

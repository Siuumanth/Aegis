package proxy

import (
	"Aegis/internal/resp"
	"context"
	"io"
	"log"
	"net"
	"time"
)

// Conn handles a single client connection lifecycle
type Conn struct {
	conn         net.Conn
	parser       *resp.Parser
	router       *Router
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewConn(conn net.Conn, router *Router, parser *resp.Parser, readTimeout *time.Duration, writeTimeout *time.Duration) *Conn {
	return &Conn{
		conn:         conn,
		parser:       parser,
		router:       router,
		readTimeout:  *readTimeout,
		writeTimeout: *writeTimeout,
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
		_ = c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))

		// 3. Parse request
		// Here, the tcp connecotin is blocked until a command is received
		cmd, err := c.parser.Parse()
		if err != nil {
			// client disconnected or read error, exit loop
			//if the client actually disconnects, the Read function will immediately stop waiting and return a specific error (io.EOF)
			if err == io.EOF {
				// normal disconnect
				return
			}

			// to handle client disconnection silently
			_, ok := err.(net.Error)
			if ok {
				return
			}

			log.Printf("parse error: %v\n", err)
			return
		}

		// 4. Route request with request context
		if err := c.router.Route(ctx, cmd, c.conn); err != nil {
			// if write fails → client likely gone
			log.Printf("routing error: %v\n", err)

			// optional: break instead of continue
			return
		}
	}
}

/*
-> more notes in tcpNreader
*/

package proxy

import (
	"io"
	"net"
)

// directly forward bytes
// for unknown commands, pipe raw bytes directly
func Passthrough(conn net.Conn, redisConn net.Conn, raw []byte) error {
	// write raw to redis
	_, err := redisConn.Write(raw)
	if err != nil {
		return err
	}
	// pipe response back to client
	io.Copy(conn, redisConn)
	return nil
}

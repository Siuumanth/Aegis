package resp

import (
	"fmt"
	"net"
)

// WriteString writes a RESP bulk string response
func WriteString(conn net.Conn, val string) error {
	_, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
	return err
}

// WriteNull writes a RESP null bulk string (key not found)
func WriteNull(conn net.Conn) error {
	_, err := fmt.Fprintf(conn, "$-1\r\n")
	return err
}

// WriteError writes a RESP error response
func WriteError(conn net.Conn, err error) error {
	_, err2 := fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
	return err2
}

// WriteOK writes a RESP simple string OK
func WriteOK(conn net.Conn) error {
	_, err := fmt.Fprintf(conn, "+OK\r\n")
	return err
}

// WriteInteger writes a RESP integer response
func WriteInteger(conn net.Conn, val int64) error {
	_, err := fmt.Fprintf(conn, ":%d\r\n", val)
	return err
}

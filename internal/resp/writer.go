package resp

import (
	"fmt"
	"net"
)

// TODO: instead of Fprintf, use conn.Write for faster allocations
// WriteString writes a RESP bulk string response
var nullBulkString = []byte("$-1\r\n")

// WriteNull writes a RESP null bulk string (key not found)

// func WriteNull(conn net.Conn) error {
// 	_, err := conn.Write(nullBulkString) // Zero logic, just raw bytes
// 	return err
// }
// func WriteString(conn net.Conn, val string) error {
// 	buf := make([]byte, 0, 64+len(val))

// 	buf = append(buf, '$')
// 	buf = strconv.AppendInt(buf, int64(len(val)), 10)
// 	buf = append(buf, '\r', '\n')
// 	buf = append(buf, val...)
// 	buf = append(buf, '\r', '\n')

// 	_, err := conn.Write(buf)
// 	return err
// }

// // WriteError writes a RESP error response
// func WriteError(conn net.Conn, err error) error {
// 	msg := err.Error()

// 	buf := make([]byte, 0, 64+len(msg))

// 	buf = append(buf, '-')
// 	buf = append(buf, "ERR "...)
// 	buf = append(buf, msg...)
// 	buf = append(buf, '\r', '\n')

// 	_, err2 := conn.Write(buf)
// 	return err2
// }

// // WriteOK writes a RESP simple string OK
// var okResp = []byte("+OK\r\n")

// func WriteOK(conn net.Conn) error {
// 	_, err := conn.Write(okResp)
// 	return err
// }

// // WriteInteger writes a RESP integer response
// func WriteInteger(conn net.Conn, val int64) error {
// 	buf := make([]byte, 0, 32)

// 	buf = append(buf, ':')
// 	buf = strconv.AppendInt(buf, val, 10)
// 	buf = append(buf, '\r', '\n')

// 	_, err := conn.Write(buf)
// 	return err
// }

// // write any type
// func WriteAny(conn net.Conn, val any) error {
// 	switch v := val.(type) {
// 	case string:
// 		return WriteString(conn, v)
// 	case int64:
// 		return WriteInteger(conn, v)
// 	case nil:
// 		return WriteNull(conn)
// 	case []any:
// 		// write array header
// 		buf := make([]byte, 0, 32)
// 		buf = append(buf, '*')
// 		buf = strconv.AppendInt(buf, int64(len(v)), 10)
// 		buf = append(buf, '\r', '\n')

// 		if _, err := conn.Write(buf); err != nil {
// 			return err
// 		}

// 		for _, item := range v {
// 			if err := WriteAny(conn, item); err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	default:
// 		return WriteError(conn, fmt.Errorf("unsupported type"))
// 	}
// }

func WriteNull(conn net.Conn) error {
	_, err := conn.Write(nullBulkString) // Zero logic, just raw bytes
	return err
}
func WriteString(conn net.Conn, val string) error {
	_, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
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

// write any type
func WriteAny(conn net.Conn, val any) error {
	switch v := val.(type) {
	case string:
		return WriteString(conn, v)
	case int64:
		return WriteInteger(conn, v)
	case nil:
		return WriteNull(conn)
	case []any:
		// array response
		fmt.Fprintf(conn, "*%d\r\n", len(v))
		for _, item := range v {
			WriteAny(conn, item)
		}
		return nil
	default:
		return WriteString(conn, fmt.Sprintf("%v", v))
	}
}

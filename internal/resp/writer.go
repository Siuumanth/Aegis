package resp

import (
	"fmt"
	"io"
	"strconv"
)

// TODO: instead of Fprintf, use conn.Write for faster allocations
// WriteString writes a RESP bulk string response
// WriteNull writes a RESP null bulk string (key not found)

var nullResp = []byte("$-1\r\n")

func WriteNull(w io.Writer) error {
	_, err := w.Write(nullResp)
	return err
}

func WriteString(w io.Writer, val string) error {
	var scratch [64]byte
	buf := scratch[:0]

	buf = append(buf, '$')
	buf = strconv.AppendInt(buf, int64(len(val)), 10)
	buf = append(buf, '\r', '\n')

	if _, err := w.Write(buf); err != nil {
		return err
	}

	if _, err := w.Write([]byte(val)); err != nil {
		return err
	}

	_, err := w.Write([]byte("\r\n"))
	return err
}

// WriteError writes a RESP error response
func WriteError(w io.Writer, err error) error {
	var scratch [64]byte
	buf := scratch[:0]

	buf = append(buf, '-')
	buf = append(buf, "ERR "...)
	buf = append(buf, err.Error()...)
	buf = append(buf, '\r', '\n')

	_, err2 := w.Write(buf)
	return err2
}

// WriteOK writes a RESP simple string OK
var okResp = []byte("+OK\r\n")

func WriteOK(w io.Writer) error {
	_, err := w.Write([]byte("+OK\r\n"))
	return err
}

// WriteInteger writes a RESP integer response
func WriteInteger(w io.Writer, val int64) error {
	var scratch [32]byte
	buf := scratch[:0]

	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, val, 10)
	buf = append(buf, '\r', '\n')

	_, err := w.Write(buf)
	return err
}

// write any type
func WriteAny(w io.Writer, val any) error {
	switch v := val.(type) {
	case string:
		return WriteString(w, v)
	case int64:
		return WriteInteger(w, v)
	case nil:
		return WriteNull(w)
	case []any:
		var scratch [32]byte
		buf := scratch[:0]

		buf = append(buf, '*')
		buf = strconv.AppendInt(buf, int64(len(v)), 10)
		buf = append(buf, '\r', '\n')

		if _, err := w.Write(buf); err != nil {
			return err
		}

		for _, item := range v {
			if err := WriteAny(w, item); err != nil {
				return err
			}
		}
		return nil
	default:
		return WriteError(w, fmt.Errorf("unsupported type"))
	}
}

// func WriteNull(conn net.Conn) error {
// 	_, err := conn.Write(nullBulkString) // Zero logic, just raw bytes
// 	return err
// }
// func WriteString(conn net.Conn, val string) error {
// 	_, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
// 	return err
// }

// // WriteError writes a RESP error response
// func WriteError(conn net.Conn, err error) error {
// 	_, err2 := fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
// 	return err2
// }

// // WriteOK writes a RESP simple string OK
// func WriteOK(conn net.Conn) error {
// 	_, err := fmt.Fprintf(conn, "+OK\r\n")
// 	return err
// }

// // WriteInteger writes a RESP integer response
// func WriteInteger(conn net.Conn, val int64) error {
// 	_, err := fmt.Fprintf(conn, ":%d\r\n", val)
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
// 		// array response
// 		fmt.Fprintf(conn, "*%d\r\n", len(v))
// 		for _, item := range v {
// 			WriteAny(conn, item)
// 		}
// 		return nil
// 	default:
// 		return WriteString(conn, fmt.Sprintf("%v", v))
// 	}
// }

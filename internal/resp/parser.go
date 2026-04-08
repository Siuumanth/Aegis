package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser reads RESP2 protocol from a connection
type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

// Parse reads one RESP2 command from the connection
func (p *Parser) Parse() (*Command, error) {
	// read raw bytes for passthrough
	line, err := p.peekLine()
	if err != nil {
		return nil, err
	}

	var raw []byte
	var args []string

	// chooose inline or resp parser

	if len(line) > 0 && line[0] == '*' {
		raw, args, err = p.readArray()
	} else {
		raw, args, err = p.readInline()
	}

	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := &Command{
		Name: strings.ToUpper(args[0]),
		Raw:  raw,
	}

	// extract key and args
	if len(args) > 1 {
		cmd.Key = args[1]
	}
	if len(args) > 2 {
		cmd.Args = args[2:]
	}
	//	PrintCommand(cmd, os.Stdout)

	return cmd, nil
}

// peeking line for inline
func (p *Parser) peekLine() ([]byte, error) {
	line, err := p.reader.Peek(1)
	if err != nil {
		return nil, err
	}
	return line, nil
}

// inline support
func (p *Parser) readInline() ([]byte, []string, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, nil, err
	}
	raw := append([]byte{}, line...)
	// remove \r\n
	str := strings.TrimSpace(string(line))
	if str == "" {
		return nil, nil, fmt.Errorf("empty inline command")
	}

	parts := strings.Split(str, " ")
	return raw, parts, nil
}

// PrintCommand prints a command in Redis syntax
// io.Writer is an interface, anytihng can be put, conn, console, buffer , well put stdout for now
func PrintCommand(cmd *Command, w io.Writer) error {
	args := []string{cmd.Name}
	if cmd.Key != "" {
		args = append(args, cmd.Key)
	}
	args = append(args, cmd.Args...)
	if _, err := fmt.Fprintf(w, "the input command is %s\r\n", strings.Join(args, " ")); err != nil {
		return err
	}
	return nil
}

/*
CMD: SET user:1 "john"

RESP2 looks like:

*3\r\n
$3\r\n
SET\r\n
$6\r\n
user:1\r\n
$4\r\n
john\r\n
*/
// readArray reads a RESP array (the standard command format)
func (p *Parser) readArray() ([]byte, []string, error) {
	var raw []byte

	// read first line e.g *3\r\n
	// THIS IS THE MAIN BLOCKING POINT
	// BLOCKS UNTILL SOME DATA COMES OVER THE CONNECTION AND IS READ
	// ERROR IF EOF is Read
	line, err := p.readLine()
	if err != nil {
		return nil, nil, err
	}
	raw = append(raw, line...)

	if len(line) == 0 || line[0] != '*' {
		return nil, nil, fmt.Errorf("expected array, got %q", line)
	}

	// parse count
	count, err := strconv.Atoi(strings.TrimSpace(string(line[1:])))
	// get args count
	if err != nil {
		return nil, nil, fmt.Errorf("invalid array length: %w", err)
	}
	// make an empty string array
	args := make([]string, 0, count)

	for i := 0; i < count; i++ {
		bulk, bulkRaw, err := p.readBulkString()
		if err != nil {
			return nil, nil, err
		}
		raw = append(raw, bulkRaw...)
		args = append(args, bulk)
	}

	return raw, args, nil
}

// readBulkString reads a RESP bulk string e.g $3\r\nGET\r\n
func (p *Parser) readBulkString() (string, []byte, error) {
	var raw []byte

	// read length line e.g $3\r\n
	line, err := p.readLine()
	if err != nil {
		return "", nil, err
	}
	raw = append(raw, line...)

	if len(line) == 0 || line[0] != '$' {
		return "", nil, fmt.Errorf("expected bulk string, got %q", line)
	}

	length, err := strconv.Atoi(strings.TrimSpace(string(line[1:])))
	if err != nil {
		return "", nil, fmt.Errorf("invalid bulk string length: %w", err)
	}

	// read exactly length bytes + \r\n
	buf := make([]byte, length+2)
	_, err = io.ReadFull(p.reader, buf)
	if err != nil {
		return "", nil, err
	}
	raw = append(raw, buf...)

	// strip \r\n
	return string(buf[:length]), raw, nil
}

// readLine reads until \r\n and returns the line including \r\n
func (p *Parser) readLine() ([]byte, error) {
	line, err := p.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line, nil
}

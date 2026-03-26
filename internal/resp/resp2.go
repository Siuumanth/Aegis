package resp

import "bufio"

// resp parser
// final result of parsing a command

type RESP2Parser struct {
	reader *bufio.Reader
}

func NewRESP2() *RESP2Parser {
	return &RESP2Parser{
		reader: &bufio.Reader{},
	}
}

/*
* → array
$ → bulk string
default → inline
*/

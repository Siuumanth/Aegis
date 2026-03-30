package handler

import (
	"Aegis/config"
	"Aegis/internal/resp"
	"net"
)

type Request struct {
	Cmd    *resp.Command
	Policy *config.PolicyConfig // nil if no match
	Conn   net.Conn
}

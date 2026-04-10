package handler

import (
	"Aegis/config"
	"Aegis/internal/resp"
	"io"
)

type Request struct {
	Cmd    *resp.Command
	Policy *config.PolicyConfig // nil if no match
	Writer io.Writer
}

package handler

import sf "Aegis/internal/singleflight"

// handler structs holds all dependencies
type Handler struct {
	sf sf.Group
}

func NewHandler() *Handler {
	return &Handler{}
}

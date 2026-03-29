package handler

import sf "Aegis/internal/singleflight"

type Handler struct {
	// hotkeys *hotkeys.HotKeys
	//tags *tags.Tags
	sf sf.Group
}

func NewHandler() *Handler {
	return &Handler{}
}

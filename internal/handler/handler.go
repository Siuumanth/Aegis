package handler

import (
	"Aegis/internal/redis"
	sf "Aegis/internal/singleflight"
)

type Handler struct {
	// hotkeys *hotkeys.HotKeys
	//tags *tags.Tags
	redis redis.Backend
	sf    *sf.Group
}

func NewHandler(cli redis.Backend) *Handler {
	return &Handler{
		redis: cli,
		sf:    &sf.Group{},
	}
}

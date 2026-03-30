package handler

import (
	"Aegis/internal/hotkeys"
	"Aegis/internal/redis"
	sf "Aegis/internal/singleflight"
	"Aegis/internal/tags"
)

type Handler struct {
	hotkeys *hotkeys.HotKeyService
	tags    *tags.TagService
	redis   redis.Backend
	sf      *sf.Group
}

func NewHandler(cli redis.Backend) *Handler {
	return &Handler{
		redis: cli,
		sf:    &sf.Group{},
	}
}

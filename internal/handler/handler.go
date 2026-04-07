package handler

import (
	"Aegis/internal/hotkeys"
	"Aegis/internal/redis"
	sf "Aegis/internal/singleflight"
	"Aegis/internal/tags"
)

type Handler struct {
	hotkeys   *hotkeys.HotKeyService
	tags      *tags.TagService
	redis     redis.Backend
	sf        *sf.Group
	redisAddr string
}

func NewHandler(cli redis.Backend, hk *hotkeys.HotKeyService, t *tags.TagService, redisAddr string) *Handler {
	return &Handler{
		redis:     cli,
		hotkeys:   hk,
		tags:      t,
		sf:        &sf.Group{},
		redisAddr: redisAddr,
	}
}

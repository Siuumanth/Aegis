package handler

import "Aegis/internal/types"

// define get handler:
/*
GET:
    singleflight.Do(key, redis.Get)
    → redis.Get(key)
    → resp.Write(result)
    → hotkeys.Track(key)        ← async, worker pool
*/

func (h *Handler) Get(req *types.Request) error {
	return nil
}

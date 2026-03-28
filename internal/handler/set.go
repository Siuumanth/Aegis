package handler

import "Aegis/internal/resp"

/*
SET:
    policy.ResolveTTL(req.Policy, cmd.ClientTTL)
    → redis.Set(key, value, ttl)
    → tags.Register(key, policy.Tags, cmd.ATags)
    → resp.Write(OK)
*/

func (h *Handler) Set(cmd *resp.Command) {

}

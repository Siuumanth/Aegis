#### Chain will look like:

```
parse command
→ policy match
→ route to handler
        │
        ├── GET handler chain
        │     singleflight
        │     → forward to Redis
        │     → hot key increment
        │     → TTL extend if hot
        │     → return
        │
        ├── SET handler chain
        │     resolve TTL
        │     → forward to Redis
        │     → register tags
        │     → return
        │
        ├── DEL handler chain
        │     forward to Redis
        │     → cleanup tag membership
        │     → return
        │
        ├── AEGIS.* handler
        │     parse subcommand
        │     → execute (invalidate, etc)
        │     → return
        │
        └── passthrough
              pipe raw bytes to Redis
              → pipe response back
              → return
```

common parameters are:
policies




#### Features:
```
✓ RESP2 transparent proxy
      - any Redis client connects without modification
      - unhandled commands pipe raw bytes directly to Redis
      - MULTI/EXEC passthrough with interception disabled

✓ TTL policy engine
      - define TTLs by key pattern in yaml
      - min_ttl / max_ttl guardrails per pattern
      - applied automatically on SET
      - defaults for unmatched keys

✓ Tag-based invalidation
      - yaml tags → auto-registered on SET via pattern match
      - ATAG in SET command → runtime tag assignment by app
      - AEGIS.INVALIDATE TAG <tagname> → atomic DEL via Lua script
      - SREM cleanup on DEL

✓ Singleflight on GET
      - deduplicates concurrent Redis reads for same key
      - per-pattern opt-in via config
      - uses golang.org/x/sync/singleflight

✓ Hot key detection
      - per-pattern config (window, threshold, ttl_multiplier)
      - system-wide max_tracked and cleanup_interval
      - on threshold crossed → EXPIRE key * ttl_multiplier
      - counter decay on cleanup tick

✗ Thundering herd    — needs read-through fetcher, cut
✗ SWR               — needs read-through fetcher, cut
✗ Negative cache    — cut
✗ Read-through      — v2
```







**Best of both worlds — support both:**

```
yaml tags   → default membership, applied automatically on SET if pattern matches
ATAG        → runtime override, app can add extra tags beyond what yaml defines
```
gf
```
PARSE request
→ identify command
→ match policy (pattern or tag)
→ route to handler
```

---
**GET**

```
GET key
→ match policy for key
→ singleflight check
  → inflight for this key? wait and reuse result
  → not inflight? mark inflight, proceed
→ forward GET to Redis
→ unmark inflight, broadcast result to waiters
→ increment hot key counter for this key
→ if counter crossed threshold within window
  → EXPIRE key (current_ttl * ttl_multiplier)
→ return result to client
```

---
**SET**
```
SET key value [EX x]
→ match policy for key
→ determine TTL
  → client provided EX/PX?
    → policy has min_ttl/max_ttl? clamp to range
    → no policy? use client TTL as-is
  → client did not provide EX?
    → policy TTL exists? apply it
    → no policy? forward as-is, no TTL injected
→ forward SET to Redis with resolved TTL
→ if policy has tags
  → SADD tag:<tagname> key  (one per tag)
→ return OK to client
```

---

**DEL**
```
DEL key [key ...]
→ forward DEL to Redis directly
→ for each deleted key
  → find tags this key belongs to (check policy match)
  → SREM tag:<tagname> key  (clean up tag membership)
→ return result to client
```

---

**AEGIS.INVALIDATE TAG tagname**
```
AEGIS.INVALIDATE TAG <tagname>
→ SMEMBERS tag:<tagname>  → get all keys under tag
→ DEL all returned keys   → atomic via Lua script
→ DEL tag:<tagname>       → remove the tag set itself
→ return count of invalidated keys to client
```

---

**Everything else**

```
any other command (HSET, ZADD, LRANGE, SUBSCRIBE, etc)
→ raw passthrough
→ pipe bytes directly to Redis
→ pipe response bytes directly back to client
→ no interception, no policy, no tracking
```

---

One thing to note on D
EL — cleaning up tag membership (SREM) is best-effort. If it fails, the tag set has a stale reference. That's fine — on next AEGIS.INVALIDATE, SMEMBERS will return the stale key, DEL on a non-existent key is a no-op, no harm done. Lazy cleanup is acceptable here.
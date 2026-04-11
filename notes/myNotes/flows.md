**Request Flow:**

```go
main.go
  → load config
  → build RuntimeConfig
  → init redis client
  → init tags, hotkeys
  → init handler (redis, tags, hotkeys)
  → init router (config, handler)
  → start TCP listener (proxy.go)
      ↓
  new client connects
      ↓
  conn.go (per goroutine)
  → create context
  → parse loop
      ↓
  resp/parser.go
  → bytes → Command{Name, Key, Args, Raw}
      ↓
  proxy/router.go
  → policy.Match(cmd.Key) → *PolicyConfig
  → switch cmd.Name
      ↓
  handlers/
  GET  → singleflight → redis.Get → async hotkeys.Track
  SET  → ResolveTTL → redis.Set → tags.Register
  DEL  → redis.Del
  AEGIS.INVALIDATE → tags.Invalidate (Lua)
  default → passthrough raw bytes
      ↓
  resp/writer.go
  → write response back to client conn
```


```
client
  ↓
resp/parser.go
  → reads raw bytes from conn
  → returns Command{Name, Key, Args, Raw}
  ↓
proxy/router.go
  → policy.engine.Match(cmd.Key) → *PolicyConfig (nil if no match)
  → builds Request{Cmd, Policy, Conn}
  → switch cmd.Name
      GET        → handlers.HandleGET(req)
      SET        → handlers.HandleSET(req)
      DEL        → handlers.HandleDEL(req)
      AEGIS.*    → handlers.HandleAegis(req)
      default    → handlers.Passthrough(req)
  ↓
handlers/
  GET:
    singleflight.Do(key, redis.Get)
    → redis.Get(key)
    → resp.Write(result)
    → hotkeys.Track(key)        ← async, worker pool

  SET:
    policy.ResolveTTL(req.Policy, cmd.ClientTTL)
    → redis.Set(key, value, ttl)
    → tags.Register(key, policy.Tags, cmd.ATags)
    → resp.Write(OK)

  DEL:
    redis.Del(key)
    → tags.Cleanup(key)
    → resp.Write(result)

  AEGIS.INVALIDATE:
    tags.Invalidate(tagname)    ← Lua script
    → resp.Write(count)

  PASSTHROUGH:
    pipe req.Raw → redis
    → pipe response → client
```

---

**Package Structure:**

```
aegis/
├── cmd/
│   └── aegis/
│       └── main.go
│
├── config/
│   ├── config.go        # raw yaml structs
│   ├── loader.go        # yaml → Config
│   ├── defaults.go      # default constants
│   └── runtime.go       # BuildRuntimeConfig, mergeDefaults
│
├── internal/
│   ├── proxy/
│   │   ├── proxy.go     # TCP listener
│   │   ├── conn.go      # per-connection goroutine, read loop
│   │   ├── router.go    # policy match, route to handler
│   │   └── request.go   # Request{Cmd, Policy, Conn}
│   │
│   ├── resp/
│   │   ├── parser.go    # bytes → Command
│   │   └── writer.go    # results → bytes
│   │
│   ├── handlers/
│   │   ├── get.go
│   │   ├── set.go
│   │   ├── del.go
│   │   ├── aegis.go
│   │   └── passthrough.go
│   │
│   ├── policy/
│   │   ├── engine.go    # Match(key) → *PolicyConfig
│   │   └── ttl.go       # ResolveTTL, ClampTTL, ExtendTTL
│   │
│   ├── tags/
│   │   └── tags.go      # Register, Cleanup, Invalidate (Lua)
│   │
│   ├── hotkeys/
│   │   ├── hotkeys.go   # Increment, IsHot, ExtendTTL
│   │   └── worker.go    # worker pool, buffered channel
│   │
│   ├── singleflight/
│   │   └── singleflight.go
│   │
│   └── redis/
│       └── client.go    # RedisBackend interface + go-redis impl
│
├── internal/
│   └── errors/
│       └── errors.go    # ErrNoPolicy, ErrInvalidCommand etc
│
├── config.yaml
└── README.md
```

---

**Import direction:**

```
main.go
  → config, proxy

proxy/router.go
  → policy, handlers, resp

handlers/*
  → policy/ttl, tags, hotkeys, singleflight, redis, resp

policy/engine.go
  → config

tags/
  → redis

hotkeys/
  → redis

redis/
  → go-redis (external only)

errors/
  → nothing (leaf package)
```

---

Two things to fix in your current code:

`internal/` appears twice in your folder — `errors` should be inside the same `internal/`, not a second one.

`config/runtime.go` handles `BuildRuntimeConfig` — `policy/engine.go` reads from the result of that. So `main.go` calls `config.BuildRuntimeConfig` first, then passes it to `policy.NewEngine(rt)`. Engine never touches raw yaml, only the built runtime config.








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
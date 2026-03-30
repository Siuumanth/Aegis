# **Aegis: Redis-Compatible Smart Cache Proxy**

## Project Proposal

---
## **Overview**

Aegis is a Go-based, RESP2-compliant proxy that intercepts Redis traffic to enforce caching policies at the infrastructure layer. It allows teams to declaratively control TTLs, key grouping, and hot key behavior via YAML — with zero application code changes.

```
[App] --(RESP2)--> [Aegis :6379] --(TCP)--> [Redis :6380]
```

---
## **Problem Statement**

Standard Redis usage pushes caching concerns into application code, leading to:

- **Policy Fragmentation:** TTL logic, cache invalidation, and hot key handling are duplicated across every service that touches Redis.
- **Cache Stampedes:** On a cache miss, concurrent requests for the same key all hit the backend simultaneously, causing upstream spikes.
- **Manual Invalidation:** No native mechanism to group related keys and invalidate them atomically — teams resort to key-scanning or application-side tracking.

---
## **Mechanism**

- **Protocol Parsing:** Aegis implements a custom RESP2 parser using `bufio` that reads bulk string lengths and array counts to reconstruct commands with byte-level accuracy. Raw bytes are preserved for passthrough commands.
    
- **Declarative Rule Engine:** YAML policies are compiled into a runtime registry at startup. For every incoming request, Aegis performs glob-based pattern matching (via `path.Match`) on the key — first match wins. A single rule for `user:*` governs all user keys automatically.
    
- **Decoupled Async Processing:** Aegis follows a client-first model — the response is written back immediately, while tag registration and hot key tracking are handed off to background worker pools via buffered channels. Workers are capped and channels are bounded; overflow is dropped silently to prevent backpressure on the hot path.
    
- **Transparent Passthrough:** Unrecognized commands are forwarded to Redis via go-redis `Do` using the existing connection pool — no new TCP handshakes, no application breakage.
    
---
## **Key Features**

### **Request Coalescing (Singleflight)**

Wraps `x/sync/singleflight` with context cancellation via `DoChan`. Concurrent `GET` requests for the same key collapse into a single upstream fetch — subsequent callers block and share the result. Configurable per policy.

### **Tag-Based Invalidation**

Maintains a forward index (`tag → keys`) and reverse index (`key → tags`) in Redis using sets. Tag registration is async, invalidation is synchronous and atomic via a Lua script. `AEGIS.INVALIDATE <tag>` deletes all tagged keys and cleans both indexes in one round trip.

### **TTL Enforcement & Clamping**

On every `SET`, Aegis resolves the final TTL against policy bounds — clamping client-provided values between `min_ttl` and `max_ttl`, falling back to policy `ttl` if none is provided. Pure functions, no side effects.

### **Adaptive Hot Key Extension**

Tracks per-key access frequency in an in-process map with a bounded size (`max_tracked`). When a key crosses a configurable `threshold`, its TTL is extended in Redis by a `ttl_multiplier`. Re-extension is gated by `min_extend_interval` to prevent redundant `EXPIRE` calls. Counts reset on a configurable interval; stale entries are evicted when their estimated Redis expiry passes.

### **Custom SET Modifiers**

Aegis extends the `SET` command with inline modifiers parsed at the tag layer:
- `AEGIS.TAG t1 t2` — append runtime tags alongside policy tags
- `AEGIS.TAG_ONLY t1 t2` — override policy tags entirely
- `AEGIS.NOTAG` — skip tag registration for this key

---
## **Technical Specifications**

- **Runtime:** Go — goroutines and buffered channels for all async work; `sync.RWMutex` for concurrent map access.
- **Protocol:** RESP2 — custom parser with pipelining awareness; raw byte preservation for passthrough.
- **Backend:** Redis 6.x/7.x standalone, forced RESP2 via go-redis `Protocol: 2`.
- **Connection Pooling:** go-redis manages the backend pool; Aegis reuses pooled connections for all operations including passthrough.
- **Observability:** Prometheus exporter planned for hit/miss ratio, singleflight collapse rate, hot key extension count, and end-to-end latency.

---
## **Scope & Limitations**

- **Compatibility:** Standalone Redis only — Cluster and Sentinel are out of scope for v1.
- **Interception:** Full handling for `GET`, `SET`, `DEL`. Custom commands: `AEGIS.INVALIDATE`. All others pass through transparently.
- **Transactions:** `MULTI/EXEC` and Pub/Sub pass through without modification — Aegis makes no guarantees about policy enforcement inside transactions.
- **Consistency:** Tag registration and hot key tracking are async and best-effort. A crash mid-flight may leave stale index entries — these are eventually cleaned up on invalidation or key expiry.
- **Negative Caching:** Cut for v1.

---


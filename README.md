
# **Aegis — Redis-Compatible Smart Cache Proxy**

## Project Proposal
---
## **Overview**

Aegis is a drop-in Redis-compatible proxy that adds caching intelligence at the infrastructure level — with zero application code changes required.

```text
your app → Aegis → Redis
```

Point your Redis URL to Aegis. Everything works as before. Advanced caching features can then be enabled through a YAML configuration file.

---
## **Problem**
Every team using Redis ends up writing the same caching logic across services:
- thundering herd protection
- TTL management
- cache invalidation

The logic is identical, only the configuration differs.
Redis itself does not handle:
- request stampedes
- stale data handling
- grouping of keys
- miss patterns

As a result:
- logic is duplicated across services
- implementations become inconsistent
- maintaining caching behavior becomes difficult
---
## **How It Works**

Aegis operates at the protocol level using RESP2.
- Any Redis client can connect without modification
- On startup, Aegis loads a YAML config and builds a rule engine
- Each incoming command is matched against configured patterns

Behavior:
- matched commands → apply caching logic
- unmatched commands → forwarded directly to Redis

Blank config:
> Aegis behaves exactly like a transparent Redis proxy
---
## **Example Config**

```yaml
patterns:
  - match: "user:*"
    ttl: 5m
    thundering_herd:
      enabled: true
      lock_ttl: 10s
      wait_timeout: 5s
      fallback: stale
    tags: [users]
    negative_cache: true
    stale_while_revalidate: true
```

---
## **Features**
All features are opt-in and applied per key pattern.

### RESP2 Compatibility
Works with any Redis client. Unrecognized commands are passed through as raw bytes.

---
### Tag-Based Invalidation
Group related keys using tags and invalidate them together with a single operation.

---
### Thundering Herd Protection
Concurrent requests for the same key are collapsed into a single upstream operation.  
Other requests wait and reuse the result.

---
### TTL Policy Engine
Centralized TTL configuration using key patterns.  
Removes hardcoded TTL values from application code.

---
### Negative Caching
Caches null responses for a short duration to prevent repeated misses.

---
### Hot Key Detection
Identifies frequently accessed keys and applies adaptive caching policies such as TTL extension and proactive refresh.

---
### Observability
Exposes metrics like cache hits, misses, latency, and hot keys using Prometheus.

---
## **Tech Stack**
- Go — proxy server, concurrency, rule engine
- Redis — backend storage
- RESP2 — client protocol
- go-redis — Redis backend communication
- Docker — deployment
- Prometheus — metrics
---
## **Scope**
- Single-node Redis support (no cluster or sentinel initially)
- Supported commands:  
    `GET, SET, DEL, EXPIRE, TTL, MGET, MSET, EXISTS`
- Custom commands: `AEGIS.*` for tag operations
- `MULTI/EXEC` passed through without interception
- Pub/Sub and Streams not supported (connect directly to Redis)
- All features are opt-in via YAML
Blank configuration ensures zero behavioral change.

---


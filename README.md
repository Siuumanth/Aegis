# **Aegis — Redis-Compatible Smart Cache Proxy**
Aegis is a Redis-compatible proxy that externalizes caching behavior from application code into a centralized, policy-driven layer. It allows defining how caching behaves (TTL, invalidation, deduplication, etc.) without modifying application logic.

---
## **1. Dependencies**

- **Core Backend:** Go (Golang) with `net` for TCP server, custom RESP parser for Redis protocol handling, YAML-based configuration parser, and Redis client
- **Caching & Data Layer:** Redis is used both as the primary datastore and for maintaining metadata such as tag indexes
- **Resilience & Concurrency:** gobreaker for Redis protection, along with goroutines, channels, mutexes, and waitgroups for concurrent processing
- **Configuration:** YAML-based configuration system for defining global settings and policies
- **Containerization:** Docker and Docker Compose 

---
## **2. System Overview & Architecture**

```text
Client → Aegis → Redis
```

Aegis sits transparently between the client and Redis, acting as a drop-in replacement. Requests are intercepted, parsed using RESP, matched against policies, processed with caching controls, and then forwarded to Redis.

System flow diagram:  
![https://github.com/Siuumanth/Aegis/blob/main/documentation/dev-notes/images/sys.png?raw=true](https://github.com/Siuumanth/Aegis/blob/main/docs/dev-notes/images/sys.png?raw=true)

The internal flow consists of a TCP proxy layer for connections, a router for command handling, a policy engine for deciding behavior based on key patterns, and a handler that integrates Redis operations with features like tags, hot keys, and singleflight.

---
## **3. Design & Implementation**
### **Configuration & Policies**

Caching behavior is fully driven by YAML configuration instead of application code.

```yaml
defaults:
  ttl: 10s
  min_ttl: 5s

aegis:
  tags: false
  hot_keys: true
  singleflight: true
```

Policies allow fine-grained control per key pattern:

```yaml
policies:
  - name: "user-profiles"
    match:
      pattern: "user:*"
    config:
      ttl: 60s
      max_ttl: 10m
      singleflight: true
      tags: [users, profile]
      hot_key:
        enabled: true
        window: 2s
        threshold: 100
        ttl_multiplier: 3
```

---
### **Request Flow**

```text
Client → Parse → Match Policy → Apply Controls → Redis → Response
```

- GET requests go through policy lookup, optional singleflight deduplication, and hot key tracking
- SET requests apply TTL rules and trigger tag registration asynchronously
---
### **Caching Controls**

- **TTL Management:** Centralized control with optional min/max bounds and safe defaults
- **Singleflight:** Deduplicates concurrent requests for the same key to reduce backend load
- **Tags:** Allows grouping keys and performing bulk invalidation
- **Hot Keys:** Detects frequently accessed keys and extends TTL dynamically
---
### **Tag-Based Invalidation**
```text
AEGIS.INVALIDATE <tag>
```
Tags are implemented using forward (`tag:<tag>`) and reverse (`key-tags:<key>`) indexes stored in Redis.  
Registration happens asynchronously using pipelines, and invalidation is handled atomically using a Lua script.

---
### **Hot Key System**
Hot keys are tracked using window-based counters in memory. When a key exceeds a threshold within a time window, its TTL is extended in a controlled manner with rate limiting to avoid excessive Redis updates.

---
### **Concurrency & Failure Handling**
- Background worker pools handle tags and hot keys asynchronously
- Channel-based design ensures request path remains non-blocking
- Redis operations are wrapped with a circuit breaker to prevent cascading failures
- System fails fast if Redis is unavailable
---
## **4. Deployment Model**

Aegis runs as a stateless proxy and can be deployed either as a binary or using containers.

**Binary execution:**
```bash
./aegis
```

Requires `aegis.yaml` in the same directory.
**Docker setup:**

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  aegis:
    build: .
    ports:
      - "6380:6380"
    volumes:
      - ./aegis.yaml:/app/aegis.yaml
    depends_on:
      - redis
```

---
## **5. Performance & Observations**

Load test report:  
[https://github.com/Siuumanth/Aegis/blob/main/Load-test-report.md?raw=true](https://github.com/Siuumanth/Aegis/blob/main/Load-test-report.md?raw=true)

Load test Summary:  
![https://github.com/Siuumanth/Aegis/blob/main/documentation/dev-notes/images/load.png?raw=true](https://github.com/Siuumanth/Aegis/blob/main/docs/dev-notes/images/load.png?raw=true)
Aegis introduces minimal overhead as a proxy while significantly reducing backend load through request coalescing and caching controls. Asynchronous processing ensures stable latency, and non-critical events may be dropped under heavy load to preserve performance.

---
## **6. Configuration Behavior**

```text
Global → Policy → Command-level overrides
```

Global settings act as hard limits, policies define behavior per key pattern, and command-level modifiers (such as tag overrides) apply where supported.

---
## **7. Supported Commands**

Standard Redis commands such as GET, SET, and DEL are supported, with Aegis applying additional logic where configured.

Custom command:

```text
AEGIS.INVALIDATE <tag>
```

All other commands are passed through transparently.

---
## **8. Known Limitations & Future Work**
Hot key tracking is instance-local and not shared across deployments. Tag metadata is not automatically cleaned when keys expire, which may lead to Redis bloat. The system does not yet support distributed coordination, and some async events may be dropped under heavy load.

Future improvements include sharded hot key maps, automatic tag cleanup, observability support, and distributed coordination.

---
## **9. Documentation**
Full documentation of architecture, rules, logic.
[https://github.com/Siuumanth/Aegis/tree/main/documentation](https://github.com/Siuumanth/Aegis/tree/main/docs)

---

Noice.
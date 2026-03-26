FEATURES and HOW i will implement them for Aegis

- RESP2 Compatibility
- Tag-Based Invalidation
- Thundering Herd Protection
- TTL Policy Engine
- Hot Key Detection
- Retries
- Observability

```bash
aegis/
├── cmd/
│   └── aegis/
│       └── main.go
│
├── config/
│   ├── config.go
│   └── loader.go
│
├── internal/
│   ├── proxy/
│   │   ├── proxy.go
│   │   ├── conn.go
│   │   ├── router.go
│   │   └── request.go
│   │
│   ├── resp/
│   │   ├── parser.go
│   │   └── writer.go
│   │
│   ├── handlers/
│   │   ├── get.go
│   │   ├── set.go
│   │   ├── del.go
│   │   ├── aegis.go
│   │   └── passthrough.go
│   │
│   ├── policy/
│   │   ├── engine.go
│   │   └── ttl.go
│   │
│   ├── tags/
│   │   └── tags.go
│   │
│   ├── hotkeys/
│   │   └── hotkeys.go
│   │
│   ├── singleflight/
│   │   └── singleflight.go
│   │
│   └── redis/
│       └── client.go
│
├── config.yaml
└── README.md
```

---
## RESP2 compatibility and YAML parsing
RESP version = **how data is encoded and sent over the wire**
### Human:

``` bash
SET user:1 "john"
```

### RESP:
```resp
*3  
$3  
SET  
$6  
user:1  
$4  
john
```

So I gotta build my own parser for this , that parses and gets the:
- command intention
- key
- value
- other params like exp
---

##  Tag-Based Invalidation
Redis does not natively support this, we can do it by parsing and checking for a custom aegis command, like:

 Custom AEGIS._ commands_* Aegis extends RESP with its own commands:

```
AEGIS.INVALIDATE TAG users
AEGIS.TAG user:123 users
```

Dev has to explicitly call these. Breaks pure drop-in but is clean and explicit. This is how Redis modules work — RediSearch uses `FT.SEARCH`, RedisJSON uses `JSON.GET`. Established convention.

```python
Standard Redis commands → Aegis passes through or applies policy → Redis AEGIS.* commands → Aegis handles entirely → never reaches Redis
```


```python
redis.Do("AEGIS.INVALIDATE", "TAG", "users") 

The go-redis client doesn't care what the command is. It just serialises whatever you give it into RESP and sends it over TCP: 

*3\r\n 
$16\r\n 
AEGIS.INVALIDATE\r\n 
$3\r\n 
TAG\r\n 
$5\r\n 
users\r\n
```

---


## Thundering Herd Protection
Prevents **duplicate work for same key happening at the same time**

Without singleflight:
```
100 requests → cache miss  
→ 100 DB calls ❌
```
With singleflight:

```
100 requests → cache miss  
→ 1 DB call ✅  
→ others wait and reuse result
```

Use GO singleflight for this, when fetching any data


---


## TTL Policy Engine
We can do it like, in the yaml, if the dev has specified initial ttls and stuff, we will use it and refresh it automatically, else we will just not override the specified ttl by the dev

Params:
- tag based ttl
- ?
---

## Negative caching 
cache nil responses, prevent repeated DB hits for missing keys
###### Implementation:
Option 1 (clean):
store special marker:  
__nil__
with short TTL (e.g. 5s)
### Behavior:
GET key → miss  
→ store __nil__ with TTL  
→ next requests don’t hit DB

👉 Add:
- separate TTL for negative cache

---
## Hot Key Detection
track **how frequently each key is being accessed in a short time window**.
Example:

- key `user:123` gets 5 requests/sec → normal
- key `feed:global` gets 500 req/sec → hot

---
##### How to implement (conceptually)
1. **Count accesses per key**
    - every GET increments a counter
2. **Use a time window**
    - reset counters every ~1 second (or sliding window if you want better)
3. **Set a threshold**
    - if a key crosses something like 100 req/sec → mark as hot

```yaml
hot_keys:
  enabled
  window
  threshold
  max_tracked_keys
  cleanup_interval

  actions:
    extend_ttl
    ttl_multiplier
    swr_enabled
    prefetch
```

Hot keys trigger adaptive caching policies like extended TTL, proactive refresh, and aggressive stale serving

---




#### Observability
Implement prometheus n stuff
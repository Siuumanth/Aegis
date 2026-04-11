# **Aegis Configuration — System File**

## **Overview**
The system configuration file defines the runtime behavior of Aegis, including networking, Redis connectivity, and global feature controls.

All requests pass through this configuration layer before any policy-specific overrides are applied.

---
## **Configuration Structure**

```text
server    → Proxy network and I/O behavior  
redis     → Backend Redis connection configuration  
defaults  → Base caching parameters (used as fallback)  
aegis     → Global feature toggles  
hot_keys  → System-wide hot key tracking configuration  
```

---
## **1. Server Configuration**

Controls how Aegis exposes itself as a Redis-compatible TCP proxy.

```yaml
server:
  host: "0.0.0.0"
  port: 6380
  read_timeout: 60s
  write_timeout: 5s
```
### Fields
- **host**  
    Network interface Aegis binds to.  
    `0.0.0.0` allows external access from all interfaces.
- **port**  
    Port on which Aegis listens for incoming connections.
- **read_timeout**  
    Maximum time to wait for a client request.
- **write_timeout**  
    Maximum time allowed to send a response to the client.
### Notes
- Protects against slow or stuck clients.
- Should be tuned based on expected latency and load.
---
## **2. Redis Configuration**

Defines how Aegis communicates with the backend Redis instance.

```yaml
redis:
  address: "localhost:6379"
  pool_size: 100
  min_idle_conns: 10
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  max_retries: 1
```
### Fields
- **address**  
    Redis endpoint in `host:port` format.
- **pool_size**  
    Maximum number of active connections in the pool.
- **min_idle_conns**  
    Minimum number of idle connections maintained.
- **dial_timeout**  
    Timeout for establishing new connections.
- **read_timeout**  
    Maximum time to wait for Redis response.
- **write_timeout**  
    Maximum time allowed to send data to Redis.
- **max_retries**  
    Retry attempts for failed operations.
### Notes
- Pool size should match expected concurrency.
- Lower retries + tighter timeouts lead to faster failure handling.
---
## **3. Defaults Configuration**
Defines baseline caching behavior when no policy override is applied.

```yaml
defaults:
  ttl: 10s
  min_ttl: 5s
  max_ttl: 20s
```
### Fields
- **ttl**  
    Default time-to-live applied to keys.
    0 TTL means no limits
- **min_ttl** , max_ttl
    Minimum and Maximum allowed TTL to prevent excessively short-lived entries.
### Notes
- Used as a fallback when no matching policy exists.
- Ensures consistent behavior across all keys.

---
## **4. Aegis Global Controls**

Controls feature enablement across the entire system.

```yaml
aegis:
  tags: false
  hot_keys: true
  singleflight: true
```

### Fields
- **tags**  
    Enables tag-based invalidation globally.
- **hot_keys**  
    Enables hot key detection and adaptive TTL logic.
- **singleflight**  
    Enables request coalescing to deduplicate concurrent requests.
### Notes
- Acts as a **global kill switch layer**.
- If a feature is disabled here, it is ignored even if enabled in policies.
---
## **5. Hot Key System Configuration**

Defines system-wide constraints and behavior for hot key tracking.

```yaml
hot_keys:
  max_tracked: 10000
  cleanup_interval: 10s
  # overridable defaults
  stale_after: 1m
  min_extend_interval: 30s
  window: 2s
  threshold: 200
  ttl_multiplier: 200
```
### Fields
- **max_tracked**  
    Maximum number of keys tracked globally (memory bound).
- **cleanup_interval**  
    Frequency of background cleanup for stale entries.
- **stale_after**  
    Duration after which an inactive key is considered stale.
- **min_extend_interval**  
    Minimum time between consecutive TTL extensions.
- **window**  
    Time window used for request rate measurement.
- **threshold**  
    Request count within the window required to classify a key as hot.
- **ttl_multiplier**  
    Factor by which TTL is extended for hot keys.
### Notes
- These act as **global defaults** for hot key behavior.
- Can be overridden at the policy level when enabled.
- Designed to prevent excessive memory usage and uncontrolled TTL growth.
---
### **Default Behavior**
- If no policy matches a key, Aegis falls back to **default configuration safely and deterministically**.
- By default:
    - **singleflight is enabled**
    - **hot key handling is disabled**
    - **tag-based invalidation is disabled**

This ensures minimal overhead while still preventing duplicate backend requests.

---
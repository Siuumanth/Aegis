# **Hot Key System — Documentation**

## **Overview**
The hot key system in Aegis dynamically detects frequently accessed keys and adjusts their TTL to reduce backend load.

Instead of relying on static TTL values, Aegis observes request patterns in real time and extends the lifetime of keys that exhibit high access frequency. This helps:
- Reduce repeated recomputation for popular keys
- Lower load on upstream systems
- Stabilize latency under bursty traffic
Hot key handling is **policy-driven**, with global defaults and per-policy overrides.

---
## **How It Works**
At a high level, the system operates in three stages:
### **1. Event Tracking (Asynchronous)**
Each request for a key generates a tracking event:
- Events are pushed into a buffered channel
- Worker goroutines consume these events
- Processing is **non-blocking** and does not impact request latency
If the system is under heavy load and the channel is full, events may be dropped (intentional in v1 to preserve performance).
---
### **2. Frequency Detection (Window-Based Counting)**
Each key maintains a lightweight in-memory entry:
- Request count within a time window
- End of the current window
- Timestamp of last TTL extension

For every event:
- If the current time is outside the window → reset count
- Otherwise → increment count

A key is classified as **hot** when:

```text
request_count ≥ threshold (within the window)
```

---
### **3. TTL Extension (Controlled Adaptation)**
When a key becomes hot:
- A TTL extension is triggered
- New TTL is derived from policy configuration
- Extension is rate-limited to avoid excessive updates

The system ensures:
- No excessive Redis calls
- No uncontrolled TTL growth
- Stable behavior under repeated bursts

---
## **Configuration Parameters**
### **Global Configuration (`hot_keys`)**
These define system-wide limits and defaults.

```yaml
hot_keys:
  max_tracked: 10000
  cleanup_interval: 10s
  stale_after: 1m
  min_extend_interval: 30s
  window: 2s
  threshold: 200
  ttl_multiplier: 200
```

---
### **Field Descriptions**
#### **max_tracked**
Maximum number of keys tracked in memory.
- Prevents unbounded memory growth
- Once limit is reached, new keys are not tracked
---
#### **cleanup_interval**
Frequency at which background cleanup runs.
- Removes stale entries
- Keeps memory usage bounded
---

### Policy specific and overridable fields:
#### **stale_after**
Duration after which a key is considered inactive.
- If no TTL extension has occurred within this time, the key is removed from tracking
- Prevents long-term retention of cold keys
---
#### **min_extend_interval**
Minimum time between successive TTL extensions for the same key.
- Acts as a **cooldown mechanism**
- Prevents repeated Redis updates for the same hot key
---
#### **window**
Time window used for measuring request frequency.
- Defines how quickly a key must accumulate requests to be considered hot
- Smaller window → more sensitive detection
---
#### **threshold**
Minimum number of requests within the window required to classify a key as hot.
- Higher values reduce false positives
- Lower values increase sensitivity
---
#### **ttl_multiplier**
Factor used to compute the extended TTL.
```text
new_ttl = policy_ttl × ttl_multiplier
```
- Allows dynamic scaling of TTL for hot keys
- Controlled to avoid excessive extension
---
## **Policy-Level Overrides**
Each policy can override hot key behavior:

```yaml
hot_key:
  enabled: true
  window: 2s
  threshold: 100
  ttl_multiplier: 3
```
### Fields
- **enabled**  
    Enables hot key logic for this policy.
- **window / threshold**  
    Override detection sensitivity.
- **ttl_multiplier**  
    Controls how aggressively TTL is extended.
- **min_extend_interval / stale_after** _(if exposed)_  
    Fine-tune stability and cleanup behavior.

---
## **Behavioral Guarantees**
- Hot key tracking is **non-blocking and asynchronous**
- TTL extension is **rate-limited and controlled**
- Memory usage is **bounded via max_tracked + cleanup**
- System degrades gracefully under load (event drops allowed)
---
## **Internal Implementation (Brief)**
The hot key system is implemented as an in-memory service with the following components:
### **Core Structures**
- A map:
    ```text
    key → {count, windowEnd, lastIncreased, policy}
    ```
- A buffered channel for event ingestion
- A worker pool for asynchronous processing
---
### **Execution Flow**
1. Requests call `Track()` → enqueue event
2. Workers consume events → update counters (`increment`)
3. If threshold is reached → trigger `Extend()`
4. TTL is updated in Redis using `EXPIRE`
5. Background cleanup removes stale entries
---
### **Concurrency Model**
- `sync.RWMutex` protects shared map access
- Workers operate independently on events
- Redis calls are performed **outside critical sections**
---
### **Design Notes**
- Window-based counting avoids continuous counter resets
- `lastIncreased` prevents excessive TTL updates
- Cleanup is time-based, not event-based
---
## **Future Improvements**
- Replace the global map + mutex with a **sharded map** to reduce lock contention under high concurrency
- Improve TTL extension logic to consider current TTL instead of static policy TTL
- Introduce adaptive thresholds based on traffic patterns
- Add observability (metrics for hot key detection and extensions)
---

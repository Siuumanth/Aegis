# **Tag System**

## **Overview**
The tag system enables grouping of keys under logical labels and supports bulk invalidation of those keys.
Each key can belong to multiple tags, and each tag can reference multiple keys.

---
## **Tag Assignment**
Tags are applied during write operations (e.g. `SET`) using:
- Policy-defined tags
- Optional command-level overrides
### **Modes**

- **Default (`AEGIS.TAG`)**  
    Merges policy tags with provided tags
    
- **Override (`AEGIS.TAG_ONLY`)**  
    Uses only provided tags
    
- **Skip (`AEGIS.NOTAG`)**  
    Disables tagging for the request
    
Example:
```r
SET users:2 chubs AEGIS.NOTAG
SET users:2 chubs AEGIS.TAG_ONLY premium-user user
SET users:2 chubs AEGIS.TAG premium-user
```
---
## **Command — Tag Invalidation**

```r
AEGIS.INVALIDATE [tag...]

eg:
AEGIS.INVALIDATE users premium-users
```
### **Behavior**
- Deletes all keys associated with the given tag(s)
- Supports multiple tags
- Returns total number of keys deleted
---
## **Client Usage — Direct Command Support**

All Aegis-specific commands and modifiers are designed to be used directly through standard Redis clients.

No custom SDK or integration layer is required.

## **Execution via Redis Client**

Clients can invoke Aegis commands using normal command execution methods (e.g. `Do`, raw command strings, or equivalent).

Example:
```go
client.Do("AEGIS.INVALIDATE", "users")
```
or in redis-cli:
```go
AEGIS.INVALIDATE users
```

---
## **Inline Modifiers in Standard Commands**
Aegis modifiers for tagging are passed as additional arguments within standard Redis commands:

```r
SET user:1 data AEGIS.TAG profile
```

```r
SET user:1 data AEGIS.TAG_ONLY custom
```

```r
SET user:1 data AEGIS.NOTAG
```

---
## **Behavior**
- Commands are parsed and intercepted by Aegis at the proxy layer
- Non-Aegis commands are passed through transparently to Redis
- Aegis-specific commands are handled internally

---
## **Key Property**
Aegis remains fully **Redis-compatible**, allowing:
- Existing clients to work without modification
- Direct command execution using standard Redis APIs
- Seamless integration into existing systems

---




---
# **Internal Model**
The system maintains **bidirectional indexing in Redis**.
### **1. Forward Index**

```text
tag:<tag> → Set of keys
```

Example:
```text
tag:users → {user:1, user:2}
```
---
### **2. Reverse Index**
```text
key-tags:<key> → Set of tags
```

Example:
```text
key-tags:user:1 → {users, profile}
```

---
## **Write Path (Tag Registration)**
- Tag resolution is performed (policy + command modifiers)
- Event is pushed asynchronously
- Worker executes a **pipeline**:

```go
SADD tag:<tag> <key>
SADD key-tags:<key> <tag>
```

- Both indexes are updated in a single network round-trip
---
## **Delete Path (Key Removal)**
When a key is deleted:
1. Fetch tags from `key-tags:<key>`
2. Remove key from all `tag:<tag>` sets
3. Delete reverse index and key
- Executed using a **pipeline** for efficiency
---
## **Invalidation Path (Core Logic)**
Tag invalidation is implemented using a **Lua script for atomic execution**
### **Flow**
1. Fetch all keys under the tag:
    ```text
    SMEMBERS tag:<tag>
    ```
    
2. For each key:
    - Fetch associated tags from reverse index
    - Remove key from all forward indexes
    - Delete:
        - Actual key
        - Reverse index
3. Delete the original tag set
---
### **Key Property**
- Entire operation is **atomic**
- Ensures:
    - No dangling references
    - No partial cleanup
    - Consistent state across indexes
---
## **Concurrency Model**
- Tag operations are asynchronous:
    - Registration → `registerChan`
    - Deletion → `deleteChan`
- Worker pool processes events
- Failures are non-blocking (do not affect client response)
---
## **Design Characteristics**
- Dual indexing enables efficient lookup in both directions
- Pipeline reduces network overhead for multi-tag writes
- Lua script ensures correctness during bulk invalidation
- System trades strict consistency for performance on write path (async model)

---

# **Trade-offs & Limitations (v1)**
- **Stale tag references on TTL expiry**  
    When a key expires naturally (TTL), it is **not automatically removed** from tag indexes.  
    This can leave stale entries in:
    
    ```text
    tag:<tag>
    key-tags:<key>
    ```
    
- **Impact**
    - Gradual growth of unused metadata in Redis
    - Slightly higher memory usage over time
    - Extra cleanup work during invalidation
- **Current Status**  
    This is a known limitation in v1 due to the asynchronous and decoupled design.
- **Recommendation**  
    Use tagging selectively for:
    - High-value grouped invalidation
    - Not for extremely high-churn, short-TTL keys
        
- **Future Fix**  
    Planned improvements include:
    - Automatic cleanup on key expiry (e.g. keyspace notifications or background reconciliation)
    - More efficient index maintenance to prevent metadata buildup
        
---
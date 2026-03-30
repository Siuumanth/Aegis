Plan of tags: 

```
SET user:123:session value ATAG session:123
SET user:123:prefs value ATAG session:123
SET user:123:cart value ATAG session:123

# user logs out
AEGIS.INVALIDATE TAG session:123
```

## 1. System Overview

The tag system in Aegis is designed to support:

```text
Group-based invalidation of cache keys
```

Instead of deleting keys individually, keys can be grouped under **tags**, and invalidation can be performed at the tag level.

---




## 2. Core Data Model
The system maintains **two indices** to ensure correctness and avoid memory leaks.
### 2.1 Forward Index (Tag → Keys)

```text
tag:<tag> → Set of keys
```

Example:
```text
tag:users → [user:1, user:2]
```

**Purpose:**
- Used during invalidation
- Allows fetching all keys under a tag
---
### 2.2 Reverse Index (Key → Tags)

```text
key-tags:<key> → Set of tags
```

Example:
```text
key-tags:user:1 → [users, profile]
```

**Purpose:**
- Used during deletion
- Allows cleaning tag sets when a key is removed
---
## Why both are needed

```text
Forward only → memory leak (stale keys stay forever)
Reverse only → cannot invalidate efficiently
```

So:
```text
Forward + Reverse = correctness + bounded memory
```

---



## 3. Write Path (SET Flow)

### Step 1: Request comes in

A SET command is processed by the handler.
At this point:
- Key is known
- Tags come from:
    - policy config
    - optional user-provided tags (ATAG)
---
### Step 2: Event is enqueued
Instead of writing tags synchronously:

```text
TagService.Register(...) pushes event into channel
```

This is done using a **non-blocking send**:
- If buffer is full → event is dropped (v1 tradeoff)
---
### Step 3: Worker processes event
A background worker consumes events and performs:
### For each tag:
1. Add key to tag set (forward index)
2. Add tag to key set (reverse index)
---

### What this means logically

```text
tag:users      += user:1
key-tags:user:1 += users
```

---
## Why async?

```text
- SET latency stays low
- Tagging does not block client
```

---




## 4. Delete Path (DEL Flow)

## Step 1: Delete event enqueued

When a key is deleted:

```text
TagService.Delete(key)
```

This is also asynchronous.

---
## Step 2: Worker processes deletion

The worker performs:
### 1. Fetch tags of the key

```r
SMEMBERS key-tags:<key>
```

This gives all tags associated with the key.

---
### 2. Remove key from all tag sets
For each tag:

```r
Remove key from tag:<tag>
```
---
### 3. Delete reverse index

```text
Delete key-tags:<key>
```

---
### Optimization used

Instead of multiple network calls:

```text
Pipeline batches all operations into one round trip
```

---
### Result

```text
No stale entries in tag sets
```

---





## 5. Invalidation Path (Tag-based Deletion)

---
### Step 1: Request

```text
Invalidate(tag)
```

---
### Step 2: Lua script execution
A Lua script is used to ensure **atomicity**.

---
### What the script does

### 1. Fetch all keys under the tag
```text
SMEMBERS tag:<tag>
```

### 2. For each key
```text
Remove tag from reverse index
```

This ensures reverse consistency.

### 3. Delete actual keys
```text
DEL key1 key2 key3
```

### 4. Delete tag set
```text
DEL tag:<tag>
```

---

### Why Lua?

```text
- Prevents race conditions
- Ensures consistency
- Runs atomically inside Redis
```

---


## 6. Concurrency Model

## Channels used

```text
registerChan → for SET
deleteChan   → for DEL
```

---

### Worker pool

Multiple workers run:
```text
for each worker:
    wait for events
    process register/delete
```

---

## Benefits

```text
- Parallel processing
- No blocking on main request path
- Scales with load
```

---



## 7. Failure Behavior

### Event drop
```text
If channel is full → event dropped
```

Effect:
```text
Possible inconsistency
Accepted tradeoff for v1
```

---
### Redis failure
```text
If Redis fails → both data + tags fail
```
No special handling needed.

---
## Crash during processing
```text
Partial updates possible
```
System recovers eventually via future operations.

---



## 8. Memory Characteristics

## Without reverse index

```text
tag sets grow forever (unbounded)
```

---
## With reverse index

```text
- extra metadata per key
- but no stale entries
- bounded memory growth
```

---
## Tradeoff

```text
+ correctness
+ predictability
- extra memory (~10–30%)
```

---

## 9. Key Design Decisions

---
#### 1. Async tagging
```text
Chosen for performance over strict consistency
```

---
#### 2. Forward + Reverse index
```text
Chosen for correctness and bounded memory
```

---
#### 3. Lua for invalidation
```text
Chosen for atomic operations
```
---
#### 4. Pipeline for deletion
```text
Chosen to reduce network overhead
```

---



## 10. End-to-End Flow Summary

## SET

```text
SET key
→ enqueue tag event
→ worker updates forward + reverse index
```

---

## DEL

```text
DEL key
→ enqueue delete event
→ worker cleans forward index + deletes reverse index
```

---
## INVALIDATE

```text
Invalidate(tag)
→ Lua:
    fetch keys
    clean reverse index
    delete keys
    delete tag
```

---

# Final Summary

```text
The system maintains bidirectional mappings between keys and tags,
uses asynchronous workers for updates, and ensures atomic invalidation
via Lua scripts, achieving correctness with bounded memory growth.
```

---


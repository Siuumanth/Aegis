# **Pattern Matching**
## **Overview**
Pattern matching determines which policy is applied to a given key.
Each policy defines a `match.pattern`, and Aegis selects the first policy whose pattern matches the key.

---
```yaml
policies:
  - name: "user-profiles"
    match:
      pattern: "user:*"
    config:
      ttl: 60s
      max_ttl: 10m
      singleflight: true
      tags: [users,profile]
      hot_key:
        enabled: true
        window: 2s
        threshold: 100
        ttl_multiplier: 3
  
  - name: "feed-cache"
    match:
      pattern: "feed:*:events"
    config:
      ttl: 15s
      singleflight: true
      tags: [feed]
      hot_key:
        enabled: true
        window: 1s
        threshold: 500
        ttl_multiplier: 2
```

## **Pattern Format**
Patterns use a simple prefix-based wildcard syntax.
```yaml
match:
  pattern: "user:*"
```
### **Rules**
- `*` is the only supported wildcard
- `*` matches **any sequence of characters (including empty)**
- Wildcards are typically used as a **suffix**

---
## **Supported Matching Behavior**

### **1. Prefix Matching (Primary Use Case)**

```yaml
pattern: "user:*"
```
Matches:
```text
user:1  
user:123  
user:profile:42  
```

Does not match:
```text
admin:user:1  
usr:1  
```
---
### **2. Exact Matching**
```yaml
pattern: "config:global"
```

Matches only:
```text
config:global
```
---
### **3. Full Wildcard (Catch-All)**
```yaml
pattern: "*"
```

Matches all keys.
Useful as a fallback policy.

---
## **Policy Resolution Rules**
- Policies are evaluated **top to bottom**
- The **first matching pattern wins**
- No further policies are evaluated after a match
---
## **Best Practices**
- Place **more specific patterns before broader ones**
    
    ```yaml
    - pattern: "user:profile:*"   # more specific
    - pattern: "user:*"           # broader
    ```
- Avoid overlapping patterns unless ordering is intentional
- Use a catch-all (`*`) only as the last policy
---
## **Naming Conventions (Recommended)**
- Use `:` as a logical namespace separator
- Keep prefixes consistent across the application

Examples:
```text
user:*
feed:*
session:*
```

---
## **Limitations**
- No regex support
- No multiple wildcards within a pattern
- No suffix-only or mid-string matching

This keeps matching fast and predictable.

---

## **Policy Configuration — TTL, Singleflight, Tags**

This section defines the configurable fields within a policy that control caching duration and request behavior.

---
## **1. TTL Configuration**

```yaml
config:
  ttl: 60s
  min_ttl: 5s
  max_ttl: 10m
```
### **Fields**
- **ttl**  
    Base time-to-live for keys matching the policy.
- **min_ttl**  
    Lower bound for TTL. Prevents very short-lived entries.
- **max_ttl**  
    Upper bound for TTL. Caps how long a key can live, including any extensions.
---
### **Behavior**
- TTL values are enforced within the range:
    
```text
min_ttl ≤ effective_ttl ≤ max_ttl
```
- If only `ttl` is provided, it is used directly
- If bounds are defined, TTL is clamped within the range
---
### **Special Case**
- A value of **`0`** means **no limit**
- If an `EXP` is added in the incoming command, it is respected and not modified.

Examples:
- `ttl: 0` → key does not expire
- `max_ttl: 0` → no upper bound
- `min_ttl: 0` → no lower bound

---
## **2. Singleflight**

```yaml
config:
  singleflight: true
```
### **Behavior**
- Deduplicates concurrent requests for the same key
- Only one request is forwarded to Redis
- Other requests wait and reuse the same result
---
## **3. Tags**

```yaml
config:
  tags: [users, profile]
```
### **Behavior**
- Associates keys with one or more tags
- Enables grouped invalidation of related keys
- Used by tag-based invalidation commands

---

Tags explained in more detail in tags.md
## **1. Configuration Precedence**

Configuration in Aegis is applied in a strict order:

```text
Global Config → Policy Config → Command-Level Overrides
```
### **Rules**
- **Global configuration has highest authority**
    - If a feature is disabled globally, it cannot be enabled anywhere else
    - Example: `aegis.tags = false` → all tagging is ignored
        
- **Policy configuration overrides defaults**
    - TTL, hot key settings, and singleflight behavior are defined per policy
    - If no policy matches, defaults are used
        
- **Command-level overrides apply last (only where supported)**
    - Applicable mainly to tagging (`AEGIS.TAG`, `AEGIS.TAG_ONLY`, `AEGIS.NOTAG`)
    - These modify behavior within the bounds of global config
---
## **2. Supported Commands & Scope**
### **Supported Core Commands**
- Standard Redis commands such as:
    - `GET`, `SET`, `DEL`, `EXPIRE`
- These are intercepted and may have Aegis logic applied (TTL, tags, etc.)
---
### **Custom Commands**
```text
AEGIS.INVALIDATE <tag> [tag...]
```

- Used for tag-based invalidation
- Handled internally by Aegis
---
### **Pass-through Behavior**
- Any unsupported or unmodified command is forwarded directly to Redis
- Response is returned transparently

---
### **Scope Limitations**

- No Pub/Sub support in v1
- No streaming or subscription-based features
- No transaction awareness (e.g. MULTI/EXEC not handled specially)
---
## **3. Known Limitations**
- **Hot key tracking is local to instance**
    - Not shared across multiple Aegis instances
    - Can lead to inconsistent behavior in distributed setups
        
- **Event drops under load**
    - Async channels (hot keys, tags) may drop events when full
    - System prioritizes request latency over strict tracking accuracy
        
- **No distributed coordination**
    - No synchronization between multiple Aegis instances
    - Each instance operates independently
- **Limited protocol awareness**
    - Advanced Redis features are passed through without optimization or control

---
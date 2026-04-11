### How hot keys service works

GET comes in → handler fires `Track(key, policy)` → enqueued in buffered channel → worker picks it up → increments count in map → if count crosses threshold → spawn goroutine to call `Extend`


**Map entry lifecycle:**
1. first GET on key → entry created with `Count: 1`
2. map full (`maxKeys`) → new keys dropped silently
3. count keeps incrementing on every GET
4. count crosses `threshold` → `Extend` fires
---
**Extend logic:**
- check `minExtendInterval` — if extended recently, skip
- `newTTL = policyTTL * multiplier` → fire `EXPIRE key newTTL` to Redis
- update `LastIncreased`
---
**Cleanup ticker (every `cleanupInterval`):**
- iterate map
- `time.Since(LastIncreased) > staleAfter` → delete entry (key went cold, probably expired in Redis too)
- otherwise → reset `Count = 0`, key has to re-earn hot status next window
---
**Edge cases:**
- channel full → event dropped silently, count just doesn't increment that once, fine
- key never extended (LastIncreased is zero) → gets evicted on first cleanup tick since `time.Since(zero)` is huge — acceptable for v1
BUT: if time is 0 , its ok evict, cuz if its hot then we will get many get req and itll get added auto to the map after the cleanup interval right
- key is hot every window → count resets to 0, but immediately climbs back to threshold and extends again next window
- concurrent GETs → all go into channel, workers serialize via mutex, no races
---


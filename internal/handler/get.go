package handler

// define get handler:
/*
GET:
    singleflight.Do(key, redis.Get)
    → redis.Get(key)
    → resp.Write(result)
    → hotkeys.Track(key)        ← async, worker pool
*/

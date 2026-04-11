## **Main Execution Flow**

## **Overview**

The `main` function is responsible for initializing all core components of Aegis and starting the TCP proxy server. It wires together configuration, Redis connectivity, background services, and request handling into a running system.

The execution follows a clear sequence:

```text
Load Config → Build Dependencies → Start Services → Accept Connections → Graceful Shutdown
```

Main code:
```go
/*
1. load config
2. build redis client
3. build tags, hotkeys
4. build handler
5. build policy engine
6. build router
7. start TCP listener → on each Accept() → NewConn(conn, router) → go conn.Handle()
*/
func main() {
    // Step 1: parse yaml
    rawConfig, err := config.Load("aegis.yaml")
    if err != nil {
        panic(err)
    }
    cfg := config.BuildRuntimeConfig(rawConfig)
    //  gloabl context to to pass around, specially for async workers
    // graceful shutdown context
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()
  
    globalCtx, cancel := context.WithCancel(ctx)
  
    // Building router + dependencies
    router, hk, tag := buildRouter(cfg, rawConfig, globalCtx)
    // start server and for each connection, handle it
    if err := serve(rawConfig.Server, router, globalCtx); err != nil {
        fmt.Println("Server error:", err)
    }
    // cancel global gorutine
    cancel()
    // graceful shutdown, wait for all async workers to finish
    if hk != nil {
        hk.Stop()
    }
    if tag != nil {
        tag.Stop()
    }
  
    if hk != nil {
        hk.Wait()
    }
    if tag != nil {
        tag.Wait()
    }
    // wait - waits till all async workers return
    // stop - closes channel
    fmt.Println("AEGIS TCP Server Stopped.")
}

// SERVE:

func serve(scfg *config.ServerConfig, router *proxy.Router, globalCtx context.Context) error {
    addr := fmt.Sprintf("%s:%d", scfg.Host, scfg.Port)
    // main tcp listen cmd
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        panic("Init Error" + err.Error())
    }
    defer ln.Close()
    fmt.Println("AEGIS is listening on PORT:", scfg.Port)
  
    // close listener when context is cancelled, which unblocks Accept()
    go func() {
        <-globalCtx.Done()
        ln.Close()
    }()
    for {
        conn, err := ln.Accept() // blocking
        if err != nil {
            select {
            case <-globalCtx.Done():
                return nil // graceful shutdown
            default:
                fmt.Println("Accept error:", err)
                continue
            }
        }

        go handleConnection(conn, router, globalCtx, &scfg.ReadTimeout, &scfg.WriteTimeout)
    }
}
  

// client = connection from app
func handleConnection(conn net.Conn, r *proxy.Router, globalCtx context.Context, readTimeout *time.Duration, writeTimeout *time.Duration) {
    parser := resp.NewParser(conn)
    // get new conn
    pconn := proxy.NewConn(conn, r, parser, readTimeout, writeTimeout)
    pconn.Handle(globalCtx)
}
  
func buildRouter(cfg *config.RuntimeConfig, rawConfig *config.Config, globalCtx context.Context) (*proxy.Router, *hotkeys.HotKeyService, *tags.TagService) {

    // 1. new redis backend client
    redisCli := redis.NewClient(rawConfig.Redis)
    redisCBClient := redis.NewCBBackend(redisCli, rawConfig.Redis)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    // check if redis is up
    if err := redisCBClient.Ping(ctx); err != nil {
        log.Fatalf("Redis is down or unreachable\nExiting...")
    }
    // 2. the handler needs the client to access redis
    // 3. the router needs the policy engine and handler
    hk := hotkeys.NewHotKeyService(cfg.GlobalConfig, redisCBClient, config.DefaultHotKeyBufSize)

    tag := tags.NewTagService(cfg.GlobalConfig, redisCBClient, config.DefaultTagBufSize)
    // init tags and hot keys
    if hk != nil {
        hk.Init(globalCtx, config.DefaultHotKeyWorkers)
    }
    if tag != nil {
        tag.Init(globalCtx, config.DefaultTagWorkers)
    }
    // build router components
    h := handler.NewHandler(redisCBClient, hk, tag, rawConfig.Redis.Address) // sf initialized internally
    p := policy.NewEngine(cfg)
    // 4. create the router
    router := proxy.NewRouter(cfg, h, p)
    return router, hk, tag
}
```

---

## **1. Configuration Loading**

```go
rawConfig, err := config.Load("aegis.yaml")
cfg := config.BuildRuntimeConfig(rawConfig)
```
- Loads the YAML configuration from the working directory
- Transforms it into a runtime-friendly structure
- Separates raw config (input) from processed config (used internally)

---
## **2. Context & Graceful Shutdown Setup**

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
globalCtx, cancel := context.WithCancel(ctx)
```
### Purpose
- Listens for OS interrupt signals (`Ctrl+C`)
- Creates a shared context (`globalCtx`) used across:
    - Workers (hot keys, tags)
    - Connection handlers
### Behavior
- On interrupt:
    - Context is cancelled
    - All goroutines receive shutdown signal

---
## **3. System Initialization (`buildRouter`)**
This is where all core components are constructed.
### **Redis Layer**
```go
redisCli := redis.NewClient(rawConfig.Redis)
redisCBClient := redis.NewCBBackend(redisCli, rawConfig.Redis)
```
- Initializes Redis client
- Wraps it with a **circuit breaker layer**
- Performs a startup health check (`PING`)
- Exits immediately if Redis is unreachable
---
### **Background Services**

```go
hk := hotkeys.NewHotKeyService(...)
tag := tags.NewTagService(...)
```
- Initializes:
    - Hot key tracking service
    - Tag management service

```go
hk.Init(globalCtx, workers)
tag.Init(globalCtx, workers)
```
- Starts worker pools for async processing
- Workers listen on internal channels
---
### **Core Components**

```go
h := handler.NewHandler(...)
p := policy.NewEngine(cfg)
router := proxy.NewRouter(cfg, h, p)
```
- **Handler**  
    Executes Redis operations and integrates:
    - Tags
    - Hot keys
    - Singleflight
- **Policy Engine**  
    Matches keys to policies and provides configuration
- **Router**  
    Coordinates:
    - Request parsing
    - Policy lookup
    - Command execution
---
## **4. TCP Server Startup**

```go
serve(rawConfig.Server, router, globalCtx)
```
### Listener Setup

```go
ln, err := net.Listen("tcp", addr)
```
- Opens a TCP socket on configured host and port
- Accepts Redis-compatible client connections

---
### Connection Handling Loop

```go
for {
    conn, _ := ln.Accept()
    go handleConnection(...)
}
```
- Each connection is handled in a separate goroutine
- Ensures concurrent request processing
---
## **5. Per-Connection Handling**
```go
parser := resp.NewParser(conn)
pconn := proxy.NewConn(conn, router, parser, ...)
pconn.Handle(globalCtx)
```
### Flow
- RESP parser reads incoming commands
- Proxy connection processes requests through router
- Applies:
    - Policy matching
    - Caching controls
    - Redis execution
---
## **6. Graceful Shutdown**
Triggered when the global context is cancelled.
### Steps
1. Stop accepting new connections
2. Close listener (unblocks `Accept()`)
3. Stop background services:

```go
hk.Stop()
tag.Stop()
```

4. Wait for workers to finish:
```go
hk.Wait()
tag.Wait()
```
---
### Guarantees
- No abrupt termination of in-flight operations
- Background workers complete pending tasks
- System exits in a controlled manner

---
## **Key Design Points**
- **Centralized initialization** via `buildRouter` keeps setup modular
- **Context propagation** ensures consistent lifecycle management
- **Async services** are fully decoupled from request path
- **Fail-fast startup** if Redis is unavailable
    
---
## **Summary**
The `main` function orchestrates the entire lifecycle of Aegis, from configuration loading to graceful shutdown. It ensures that all components—networking, policy engine, Redis backend, and background services—are initialized correctly and operate under a shared context for coordinated execution and termination.
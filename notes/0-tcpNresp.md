# 🧠 TCP Proxy — How It Works

```go

    fmt.Println("Starting AEGIS TCP Server...")

    // main tcp listen cmd
    ln, err := net.Listen("tcp", ":6379")
    if err != nil {
        panic(err)
    }
    defer ln.Close()
  
    for {
        // for every connectoin, accept and handle
        conn, err := ln.Accept()
        if err != nil {
            continue
        }
        go handleConnection(conn)
    }
}

// client = connection from app
func handleConnection(client net.Conn) {
    defer client.Close()
  
    // accept conn, forward to redis
    redisConn, err := net.Dial("tcp", "localhost:6380")
    if err != nil {
        return
    }
    defer redisConn.Close()
    // Bidirectional forwarding
    // 1. read from redis, write to client
    go io.Copy(redisConn, client)

    // 2. read from client, write to redis
    io.Copy(client, redisConn)
}
```

## Overview
You built a **TCP proxy** that sits between:

```text
Client ↔ Proxy ↔ Redis
```

It does:
- accept client connections
- open connection to Redis
- forward data both ways

It does **not understand the data yet**.

---

# ⚙️ Core Components

## 1. TCP Listener

- binds to a port
- waits for incoming connections
    

```text
:6379 → proxy listening
```

When a client connects:
- a new connection is created
- handed to your handler
---
## 2. Accept Loop

Runs forever:

```text
while true:
    accept connection
    spawn handler (goroutine)
```

Why goroutine?
- each client must be handled independently
- TCP is blocking
---
## 3. Connection (net.Conn)

Represents a live TCP stream.
Capabilities:
- read bytes
- write bytes
- close connection

Important:
- it’s just a **byte stream**, no structure

---
# 🔁 Core Logic — handleConnection

## What happens per client

1. client connects to proxy
2. proxy connects to Redis
3. proxy links both connections
4. data flows both directions

---
## Data Flow

### Client → Redis

```text
client sends bytes
→ proxy reads
→ proxy writes to Redis
```

---

### Redis → Client

```text
Redis responds
→ proxy reads
→ proxy writes back to client
```

---

## Full Duplex Communication

TCP allows:
- both sides to send at any time
    

So you need **two independent flows**:

```text
client → Redis
Redis → client
```

This is why:
- one direction runs in a goroutine
- the other runs in main thread
---
# 🔄 Continuous Streaming

Forwarding is not “request-response” at code level.
It’s:
- continuous stream copying
- runs until connection closes
---
# 🧱 Mental Model

Your proxy is currently:
> a pipe connecting two sockets

Like:

```text
[Client Socket] ===== [Proxy] ===== [Redis Socket]
```

Proxy does:
- no interpretation
- no modification
- just relays bytes
---
# ⚠️ Important Behaviors

## 1. Blocking I/O
- reads block until data arrives
- writes block until sent

Go handles this well with goroutines.

---
## 2. Connection lifecycle
- client closes → proxy stops
- Redis closes → proxy stops
- both connections cleaned up
---
## 3. Stateless proxy (currently)

Right now:
- no memory of requests
- no caching
- no metadata

Each connection is independent.

---
# 🧪 Validation

You verified:
- `redis-cli -p 6379` works
- commands behave same as direct Redis
- GUI shows correct data

This proves:
- protocol is preserved
- no corruption
- forwarding is correct

---
# 🧠 Key Insight

At this stage:
> You are operating at **transport layer (TCP)**

NOT:
- application layer
- Redis command level
---
# 🚧 Limitation of current design

Because you’re using raw forwarding:
- you **cannot inspect commands**
- you **cannot modify behavior**
- you **cannot implement caching logic**
---
# 🔜 What changes next

To make Aegis “smart”, you must:

1. read request fully
2. parse it (RESP)
3. decide what to do
4. optionally forward

So flow becomes:

```text
client → read → interpret → act → forward → respond
```

Instead of:

```text
client → blindly forward → done
```

---
# 💯 Summary
- TCP proxy = bidirectional byte relay
- uses listener + connections + goroutines
- `io.Copy` enables continuous streaming
- no parsing, no logic yet
- acts as transparent middle layer

---
## One-line understanding

> “A TCP proxy connects two sockets and continuously forwards data between them without understanding it.”

---


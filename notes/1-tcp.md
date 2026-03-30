# 1. TCP communication (setup in main)

### What TCP gives you

```text
TCP = reliable byte stream between client and server
```

No structure, just bytes flowing.

---

### Server setup

```go
ln, _ := net.Listen("tcp", ":6379")
```

```text
- opens port 6379
- starts listening for incoming connections
```

---

### Accept connections

```go
for {
    conn, _ := ln.Accept()
    go handleConnection(conn)
}
```

```text
- Accept() → blocks until client connects
- returns net.Conn (represents that connection)
- spawn goroutine → handle multiple clients
```

---

### What is `net.Conn`

```text
conn = bidirectional stream
- implements io.Reader  (read bytes)
- implements io.Writer  (write bytes)
```

---

### Data flow at TCP level

```text
Client writes → conn.Read() on server
Server writes → conn.Write() → client receives
```

---

# 2. Per-connection lifecycle

```go
func handleConnection(conn net.Conn) {
    parser := resp.NewParser(conn)
    pconn := proxy.NewConn(conn, router, parser)
    pconn.Handle()
}
```

```text
Each connection:
- gets its own parser
- runs its own loop
```

---

# 3. Reading data (how bytes are consumed)

Inside parser:

```go
bufio.NewReader(conn)
```

```text
conn → raw TCP stream
bufio → buffered reads (efficient)
```

---

### Example read

```go
line, _ := reader.ReadBytes('\n')
```

```text
Reads until newline (\n)
Used because RESP is line-based
```

---

# 4. RESP protocol (structure on top of TCP)

TCP gives bytes → RESP gives structure.

Example:

```text
*2\r\n
$3\r\n
GET\r\n
$7\r\n
user:42\r\n
```

---

### Meaning

```text
*2 → array of 2 elements
$3 → string of length 3 → GET
$7 → string of length 7 → user:42
```

---

# 5. Parser converts bytes → Command

```go
cmd, _ := parser.Parse()
```

Result:

```text
Command{
  Name: "GET"
  Key:  "user:42"
}
```

---

# 6. Router (decision layer)

```go
router.Route(ctx, cmd, conn)
```

What it does:

```text
- match key with policy (user:*)
- build Request
- decide handler (GET, SET, etc)
```

---

# 7. Handler (execution layer)

Example SET:

```go
ttl := ResolveTTL(policy, clientTTL)
redis.Set(key, value, ttl)
conn.Write("+OK\r\n")
```

Example GET:

```go
val := redis.Get(key)
conn.Write("$len\r\nvalue\r\n")
```

---

# 8. Redis backend call

```text
Aegis → go-redis client → Redis server (:6380)
```

```text
You are NOT using raw TCP here
go-redis handles protocol internally
```

---

# 9. Writing response back

```go
conn.Write(response)
```

```text
- must be RESP format
- client expects exact protocol
```

---

# Full end-to-end flow

```text
Client
  ↓
TCP connect (net.Dial)
  ↓
Server Accept()
  ↓
net.Conn (stream)
  ↓
bufio.Reader
  ↓
RESP parser
  ↓
Command struct
  ↓
Router (policy match)
  ↓
Handler (logic)
  ↓
Redis (backend)
  ↓
RESP response
  ↓
conn.Write()
  ↓
Client receives
```

---

# Key concepts to remember

```text
TCP → just bytes
RESP → structure on top of bytes
Parser → converts bytes → command
Router → decides what to do
Handler → executes
conn.Write → sends response
```

---

# What Aegis is doing

```text
Acts as a middle layer:

Client ↔ Aegis ↔ Redis
```

```text
- speaks RESP on both sides
- adds logic (TTL, tags, policies)
```

---

If you’re clear on this, you’ve basically understood:

- how Redis protocol works
    
- how proxies work
    
- how real infra systems are built
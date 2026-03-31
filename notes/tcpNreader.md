# What happens when client disconnects? 
## **1. The Physical Layer: TCP Connection Termination**

When a Redis client (like `redis-cli` or a Go app) closes, the Operating System sends a **FIN (Finish)** packet over the network.

- **The OS Role:** The Linux/Windows kernel receives this packet and marks the specific file descriptor (the socket) as "Closed for Reading."
- **The Go Runtime Role:** Go’s network poller (the `net` package) is constantly watching these sockets. When it sees the `FIN` flag, it prepares to return a special error to any code currently waiting on that socket.

---
## **2. The Interface Layer: From Socket to Reader**

Your parser doesn't talk to the socket; it talks to a **`bufio.Reader`**, which wraps an **`io.Reader`**.
### **The Chain of Command:**
1. **`net.Conn` (The Socket):** This is the physical implementation of the `Read` method.
2. **`io.Reader` (The Interface):** This is a "contract" that says: _"I will give you bytes or an error."_
3. **`bufio.Reader` (The Buffer):** This sits on top of the interface to make reading efficient by grabbing large chunks of data at once.

**How the error travels:**
When the client disconnects, `net.Conn.Read()` returns `0` bytes and the error `io.EOF`. Because `bufio.Reader` is just a wrapper, it sees that `io.EOF` and passes it directly to whoever is calling it.

---

# HOW IO.Reader and Conn works
```text
io.Reader doesn’t know about networks
net.Conn implements io.Reader
```

That’s the whole “wiring”.

---
# What’s actually happening

## 1. `io.Reader` is just a contract

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

```text
No TCP
No sockets
No EOF logic
```

It just says:
```text
"give me bytes or an error"
```

---
## 2. `net.Conn` implements that contract

```go
type Conn interface {
    Read(b []byte) (n int, err error)
}
```

Internally:
```text
net.Conn → wraps OS socket (file descriptor)
```

So when you do:
```go
conn.Read(buf)
```

You’re actually calling:
```go
Go → runtime → OS socket → kernel
```

---
# 3. Where EOF actually comes from

This is the key:
```text
EOF is NOT created by io.Reader
EOF is returned by net.Conn (which talks to OS)
```

---
## Flow

```text
Client closes → TCP FIN packet
↓
OS marks socket as closed for reading
↓
Go runtime netpoll sees it
↓
next conn.Read() → returns (0, io.EOF)
↓
bufio.Reader → just passes it up
↓
your parser → receives EOF
```

---

# 4. Why `bufio.Reader` doesn’t interfere
```go
reader := bufio.NewReader(conn)
```

Internally:
```text
bufio.Reader → calls conn.Read()
```

If conn returns:
```go
0, io.EOF
```

bufio does:
```go
return 0, io.EOF
```
No magic.

---
# 5. Important nuance (interview-level)

EOF doesn’t always mean “connection closed immediately”

```go
EOF = "no more data will ever come"
```

Example:

```go
Client sends last bytes + FIN
Server reads remaining buffer → gets data
Next read → EOF
```

---
# 6. Why this abstraction is powerful

Because of this:
```text
Your parser works with ANY io.Reader:
- TCP connection
- file
- string buffer
- test mock
```

Example:
```go
reader := strings.NewReader("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n")
```

Same parser works.

---
# Final mental model
```js
io.Reader → interface (dumb contract)
net.Conn → real implementation (talks to OS)
OS → detects TCP close
Go runtime → converts to io.EOF
```

---
## Brutal clarity

```text
EOF is a NETWORK EVENT surfaced as a GENERIC INTERFACE ERROR
```

That’s the beauty of Go’s design.

---
## **5. Summary: Why this is "Clean" Engineering**

| **Component**       | **Responsibility** | **Action on Disconnect**                                   |
| ------------------- | ------------------ | ---------------------------------------------------------- |
| **TCP Socket**      | Network Traffic    | Receives `FIN`, returns `io.EOF`.                          |
| **`io.Reader`**     | Abstract Data      | Passes the `io.EOF` error up the chain.                    |
| **RESP Parser**     | Protocol Logic     | Detects error, stops parsing, returns error to loop.       |
| **`Handle()` Loop** | Lifecycle          | Breaks loop, hits `defer c.conn.Close()`, releases memory. |





# My Conn.Handle() Code
```go
func (c *Conn) Handle(globalCtx context.Context) {
	ctx, cancel := context.WithCancel(globalCtx)
	defer cancel()
	defer c.conn.Close()

	for {
		// 1. Respect shutdown
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 2. Set read deadline (prevents dead connections)
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// 3. Parse request
		cmd, err := c.parser.Parse()
		if err != nil {
		// client disconnected or read error, exit loop
            //if the client actually disconnects, the Read function will immediately stop waiting and return a specific error (io.EOF)
			if err == io.EOF {
				// normal disconnect
				return
			}

			// unexpected error
			fmt.Printf("parse error: %v\n", err)
			return
		}

		// 4. Route request
		if err := c.router.Route(ctx, cmd, c.conn); err != nil {
			// if write fails → client likely gone
			fmt.Printf("routing error: %v\n", err)

			// optional: break instead of continue
			return
		}
	}
}
```

---
# What this function really is
```text
This is the connection lifecycle controller
```

It owns:

```text
- socket lifetime
- request loop
- cleanup
```

---
# **Connection Lifecycle — Handle Loop**

## **Overview**
This function manages the **entire lifecycle of a single client connection**.

```text
Accept connection → Read commands → Route → Handle errors → Cleanup
```

It acts as the **controller layer** between:
```text
TCP socket ↔ RESP parser ↔ Router/Handler
```

---
## **1. Context Management**

```go
ctx, cancel := context.WithCancel(globalCtx)
defer cancel()
```
#### **Purpose**
```text
Create a connection-scoped context
```
#### **Behavior**
- Inherits from server/global context
- Cancels when:
    - connection ends
    - server shuts down
#### **Why**
```text
Ensures all downstream work can be stopped cleanly
```

---
## **2. Connection Cleanup**
```go
defer c.conn.Close()
```
### **Purpose**
```text
Guarantees socket is closed on exit
```
### **Why**
```text
Prevents file descriptor leaks (critical in high-concurrency systems)
```

---
## **3. Event Loop (Core Execution)**
```go
for {
```
### **Role**
```text
Handles multiple commands over a single TCP connection
```

Each iteration = **one Redis command**

---
## **4. Context Cancellation Check**
```go
select {
case <-ctx.Done():
    return
default:
}
```

#### **Purpose**
```text
Graceful shutdown handling
```

#### **Why**
```text
Prevents goroutine leaks when server stops
```

---
## **5. Read Deadline (Connection Safety)**
```go
c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
```

### **Purpose**
```text
Avoid infinite blocking on dead/idle connections
```

### **Behavior**
- If no data arrives → `Read()` returns timeout error
- Triggers loop exit
---
## **6. Parsing Layer (Protocol Boundary)**

```go
cmd, err := c.parser.Parse()
```
### **Flow**
```text
parser → bufio.Reader → net.Conn → OS socket
```
### **On success**
```text
Valid RESP command constructed
```

---
### **On failure**

```go
if err != nil {
```

#### Case 1: Normal disconnect
```go
if err == io.EOF {
    return
}
```

```text
Client closed connection (TCP FIN)
```

---
#### Case 2: Unexpected error
```go
fmt.Printf("parse error: %v\n", err)
return
```

```text
Malformed input / network issue
```

---
## **7. Routing Layer (Execution)**
```go
err := c.router.Route(ctx, cmd, c.conn)
```
### **Purpose**
```text
Execute command and write response
```

---
### **On failure**
```go
if err != nil {
    return
}
```

### **Why**
```text
Write failure usually means client is gone
→ no point continuing
```

---
## **8. Full Lifecycle Flow**

```python
Client connects
↓
Handle() starts
↓
Loop:
  wait for data (Parse blocks)
  ↓
  command received → route
  ↓
  response written
↓
Client disconnects OR error occurs
↓
Parse returns error (EOF or other)
↓
Function returns
↓
defer:
  cancel()
  conn.Close()
↓
Connection fully cleaned
```

---
## **9. Key Design Principles**

## **Event-Driven via Errors**
```text
No polling, no "is alive?" checks
Errors signal state changes
```

---
## **Separation of Concerns**
```text
Parser → protocol
Router → business logic
Handle → lifecycle
```

---
## **Fail-Fast Behavior**
```text
On critical error → exit immediately
```

---
## **Resource Safety**
```text
defer ensures cleanup always runs
```

---
# **10. Mental Model**
```go
Handle() = connection state machine

WAIT → PARSE → EXECUTE → REPEAT
          ↓
        ERROR → EXIT → CLEANUP
```

---
# **Final Insight**

```text
You are not managing TCP directly

You are reacting to io.Reader behavior,
which abstracts network events (like disconnects) as errors
```

---


# 1. What is `io.Reader`?

In Go, `io.Reader` is one of the most fundamental abstractions. It represents **anything you can read bytes from** — files, network connections, buffers, etc.

The interface looks like this:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

### How it works
- You pass a byte slice (`p`) to `Read`
- Go fills that slice with data
- It returns how many bytes were actually read

Important points:
- It may read **less than requested**
- You may need to call `Read` multiple times
- When no more data → returns `io.EOF`
---
### Real-world examples of `io.Reader`

- `net.Conn` → reading from a TCP connection
- `os.File` → reading from a file
- `strings.Reader` → reading from a string
- `bytes.Buffer`

In your case:

```go
func NewParser(r io.Reader)
```

This means your parser works with **any source**, not just network connections. That’s good design.
any data can come into the reader.

---
# 2. What is `bufio.Reader`?

`bufio.Reader` is a wrapper around `io.Reader` that adds **buffering and helper methods**.
### Why buffering matters

Without buffering:
- Every read → system call (slow)

With buffering:
- Reads a large chunk once
- Serves smaller reads from memory (fast)
---
### Your usage

```go
reader *bufio.Reader
```

```go
p := &Parser{reader: bufio.NewReader(r)}
```

So the flow is:

```text
net.Conn → bufio.Reader → your parser
```

---
# 3. Important `bufio.Reader` methods

### ReadBytes

```go
line, err := reader.ReadBytes('\n')
```

This reads until it sees `\n` (newline).
RESP protocol uses `\r\n`, so this works perfectly.
Example:

```
*3\r\n
```

You read the full line in one call.

---
# 4. What is `io.ReadFull`?

```go
io.ReadFull(reader, buf)
```

This function ensures:
> “Read exactly N bytes or fail”

---
### Why you need it

In RESP:

```
$4\r\n
john\r\n
```
- `$4` tells you length = 4
- So you MUST read exactly 4 bytes + `\r\n`

If you use normal `Read`, you might get partial data.
`ReadFull` guarantees correctness.

---
# 5. RESP Protocol basics

Redis uses RESP (Redis Serialization Protocol).
You are implementing RESP2.

---
### Array

```
*3\r\n
```

- `*` → array
- `3` → number of elements
---
### Bulk string

```
$4\r\n
john\r\n
```

- `$` → bulk string
- `4` → length of string

---
### Full command example

```
SET user:1 john
```

RESP form:

```
*3\r\n
$3\r\n
SET\r\n
$6\r\n
user:1\r\n
$4\r\n
john\r\n
```

---
# 6. How your parser works

### Step 1: Read array
```go
line, _ := p.readLine()
```

Reads:
```
*3\r\n
```

Then:
```go
count := 3
```

---
### Step 2: Loop and read elements

```go
for i := 0; i < count; i++
```
Each iteration:

```go
readBulkString()
```

---
### Step 3: Read bulk string

Inside `readBulkString`:

1. Read length line:
    ```
    $4\r\n
    ```
2. Parse length = 4
3. Read exact data:
    ```
    john\r\n
    ```
---
### Step 4: Build args
```go
args = ["SET", "user:1", "john"]
```
---
### Step 5: Convert to Command
```go
cmd := &Command{
    Name: "SET",
    Key:  "user:1",
    Args: ["john"],
}
```
---
# 7. Why you store `raw []byte`

You are doing:

```go
raw = append(raw, ...)
```

This captures the original request exactly.
### Why?

For passthrough:
```text
Client → Aegis → Redis
```

If command is unknown:
- Just forward raw bytes to Redis
- No parsing required again
---
# 8. Design quality

Your parser is actually well-designed because:
- It is **streaming** (doesn’t load everything at once)
- It uses **buffering** (efficient)
- It respects protocol structure
- It separates parsing from execution

---
# Final mental model

Think of it like this:

```text
TCP connection (net.Conn)
        ↓
io.Reader (generic interface)
        ↓
bufio.Reader (efficient reading)
        ↓
Parser (understands RESP)
        ↓
Command struct (usable form)
```

---



# Full flow (from scratch)

### 1. TCP connection

```go
ln, _ := net.Listen("tcp", ":6379")
conn, _ := ln.Accept()
```

```text
Client sends bytes → arrive on TCP socket
```
---
### 2. net.Conn

```go
conn net.Conn
```

```text
conn is BOTH:
- io.Reader  (can read data)
- io.Writer  (can write data)
```

---
### 3. bufio.Reader wraps it

```go
reader := bufio.NewReader(conn)
```

```text
conn (slow, system calls)
↓
bufio.Reader (buffered, fast)
```

---
### 4. Your parser

```go
parser := NewParser(conn)
```

Inside:

```go
bufio.NewReader(conn)
```

---
### 5. Reading happens

```go
line, _ := reader.ReadBytes('\n')
```

```text
bufio.Reader:
- if buffer empty → reads from TCP
- stores chunk in memory
- returns requested data
```
---
### 6. Parser builds command

```text
Raw bytes → RESP parsing → Command struct
```

---
### 7. Your system processes

```text
Command → Router → Handler → Redis
```
---
### 8. Response goes back

```go
conn.Write(responseBytes)
```

```text
Your server → TCP → client
```

---
# Simple mental flow

```text
Client
  ↓
TCP socket
  ↓
net.Conn (Reader)
  ↓
bufio.Reader (buffer)
  ↓
Parser
  ↓
Command
  ↓
Handler
  ↓
conn.Write()
  ↓
Client
```

---
# Key idea

```text
Data flow:

Client → conn → buffer → parser → your logic
```
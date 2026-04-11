## **Deployment Model**
## **Overview**
Aegis is deployed as a **stateless proxy** that operates between the client and Redis.

```text
Client → Aegis → Redis
```

It does not maintain persistent state and relies entirely on Redis for data storage and tag indexing. This allows it to be deployed flexibly across environments.

---
## **Deployment Options**

### **1. Binary Execution**
Aegis can be run as a compiled executable.

```text
./aegis
```
### **Configuration Requirement**
The configuration file (`aegis.yaml`) must be present in the **same working directory** as the executable at runtime.
- The binary does not accept a `--config` flag in the current version
- Configuration loading is implicit and path-dependent
### **Implication**
- The executable and configuration file must be co-located
- Deployment scripts or environments must ensure correct file placement
---
### **2. Docker Deployment (Recommended)**
Aegis can be deployed using Docker along with Redis.
#### **Docker Compose Setup**

`docker-compose.yaml`
```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: redis-aegis
    ports:
      - "6379:6379"

  aegis:
    build: .
    container_name: aegis
    ports:
      - "6380:6380"
    volumes:
      - ./aegis.yaml:/app/aegis.yaml
    depends_on:
      - redis
```
Make sure the aegis port is used as the port in the "redis client" in your backend app.
### **Behavior**
- Redis runs as the backend datastore
- Aegis connects to Redis using the address defined in `aegis.yaml`
- The configuration file is mounted into the container at runtime
---
## **Container Build**
Aegis uses a multi-stage Docker build to produce a lightweight runtime image.
### **Build Stage**
- Uses a Go base image to compile the binary
- Produces a statically linked executable
### **Runtime Stage**
- Uses a minimal Alpine image
- Copies only the compiled binary
- Exposes the Aegis port

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o aegis ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/aegis .

EXPOSE 6380

CMD ["./aegis"]
```

---
## **Scaling Model**

Aegis is designed to be horizontally scalable due to its stateless nature.
Multiple instances can be deployed behind a load balancer. However, certain behaviors are instance-local:
- Hot key tracking is maintained in-memory per instance
- There is no coordination between instances for hot key detection

In contrast:
- Tag metadata is stored in Redis and remains consistent across instances

---
## **Summary**

Aegis can be deployed either as a standalone binary or as a containerized service. The current implementation requires the configuration file to be present in the working directory, which simplifies loading but imposes constraints on deployment structure. The system’s stateless design enables horizontal scaling, with clearly defined trade-offs in locally managed features.
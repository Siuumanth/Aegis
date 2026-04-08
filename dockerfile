# Build, Compile
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o aegis ./cmd/main.go

# Runtime
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/aegis .

EXPOSE 6380

CMD ["./aegis"]
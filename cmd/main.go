package main

import (
	"Aegis/config"
	"Aegis/internal/handler"
	"Aegis/internal/policy"
	"Aegis/internal/proxy"
	"Aegis/internal/redis"
	"Aegis/internal/resp"
	"fmt"
	"net"
)

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
	// yaml parser
	yaml, err := config.Load()
	if err != nil {
		panic(err)
	}

	cfg := config.BuildRuntimeConfig(yaml)
	fmt.Println(cfg)

	// build dependencies
	// 1. new redis backend client
	redisClient := redis.NewClient("localhost:6380")

	// 2. the handler needs the client to access redis
	// 3. the router needs the policy engine and handler
	h := handler.NewHandler(redisClient) // TODO: add tags and hotkeys
	p := policy.NewEngine(cfg)

	// 4. create the router
	router := proxy.NewRouter(cfg, h, p)

	// start server and for each connection, handle it

	fmt.Println("Starting AEGIS TCP Server...")
	// main tcp listen cmd
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("AEGIS TCP Server started...")

	for {
		// for every connectoin, accept and handle
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, router)
	}
}

// client = connection from app
func handleConnection(conn net.Conn, r *proxy.Router) {

	parser := resp.NewParser(conn)
	// get new connection
	pconn := proxy.NewConn(conn, r, parser)

	pconn.Handle()
}

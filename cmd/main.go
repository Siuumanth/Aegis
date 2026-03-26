package main

import (
	policy "Aegis/internal/policy"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	// yaml parser
	cfg, err := policy.Load()
	if err != nil {
		panic(err)
	}
	fmt.Println(policy.BuildRuntimeConfig(cfg))

	//fmt.Println(cfg)

	time.Sleep(time.Second * 3)
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

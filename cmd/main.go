package main

import (
	"Aegis/config"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
)

// handles one client connection
func handleConn(client net.Conn, backendAddr string) {
	defer client.Close()

	backend, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Println("backend dial error:", err)
		return
	}
	defer backend.Close()

	// client → redis
	go func() {
		io.Copy(backend, client)
		backend.Close()
	}()

	// redis → client
	io.Copy(client, backend)
}

func main() {
	rawConfig, err := config.Load("aegis.yaml")
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := fmt.Sprintf("%s:%d", rawConfig.Server.Host, rawConfig.Server.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Println("BLANK PROXY listening on", addr)

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		client, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("BLANK PROXY stopped")
				return
			default:
				log.Println("accept error:", err)
				continue
			}
		}

		go handleConn(client, rawConfig.Redis.Address)
	}

	fmt.Println("BLANK PROXY stopped")
}

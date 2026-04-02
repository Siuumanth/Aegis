package main

import (
	"Aegis/config"
	"Aegis/internal/handler"
	"Aegis/internal/hotkeys"
	"Aegis/internal/policy"
	"Aegis/internal/proxy"
	"Aegis/internal/redis"
	"Aegis/internal/resp"
	"Aegis/internal/tags"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
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
	// Step 1: parse yaml
	rawConfig, err := config.Load("test.yaml")
	if err != nil {
		panic(err)
	}
	cfg := config.BuildRuntimeConfig(rawConfig)
	config.PrintRTConfig(cfg)
	//  gloabl context to to pass around, specially for async workers
	// graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	globalCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Building router + dependencies
	router := buildRouter(cfg, rawConfig, globalCtx)

	// start server and for each connection, handle it
	fmt.Printf("AEGIS listening on %s:%d\n", rawConfig.Server.Host, rawConfig.Server.Port)

	// start server
	if err := serve(rawConfig.Server, router, globalCtx); err != nil {
		fmt.Println("Server error:", err)
	}
	fmt.Println("AEGIS TCP Server Stopped.")
}

// client = connection from app
func handleConnection(conn net.Conn, r *proxy.Router, globalCtx context.Context, readTimeout, writeTimeout time.Duration) {
	parser := resp.NewParser(conn)
	// get new conn
	pconn := proxy.NewConn(conn, r, parser, readTimeout, writeTimeout)
	pconn.Handle(globalCtx)
}

func buildRouter(cfg *config.RuntimeConfig, rawConfig *config.Config, globalCtx context.Context) *proxy.Router {
	// build dependencies
	// 1. new redis backend client
	redisClient := redis.NewClient(rawConfig.Redis)
	// tihs is resolved during building runtime config

	// 2. the handler needs the client to access redis
	// 3. the router needs the policy engine and handler
	hk := hotkeys.NewHotKeyService(cfg.GlobalConfig, redisClient, config.DefaultHotKeyBufSize)
	tag := tags.NewTagService(cfg.GlobalConfig, redisClient, config.DefaultTagBufSize)
	// init tags and hot keys
	if hk != nil {
		hk.Init(globalCtx, config.DefaultHotKeyWorkers)
	}
	if tag != nil {
		tag.Init(globalCtx, config.DefaultTagWorkers)
	}

	// build router components
	h := handler.NewHandler(redisClient, hk, tag) // sf initialized internally
	p := policy.NewEngine(cfg)

	// 4. create the router
	router := proxy.NewRouter(cfg, h, p)
	return router
}

func serve(scfg *config.ServerConfig, router *proxy.Router, globalCtx context.Context) error {
	addr := fmt.Sprintf("%s:%d", scfg.Host, scfg.Port)
	// main tcp listen cmd
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("AEGIS is listening on PORT:", scfg.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-globalCtx.Done():
				return nil // graceful shutdown
			default:
				fmt.Println("Accept error:", err)
				continue
			}
		}

		go handleConnection(conn, router, globalCtx, scfg.ReadTimeout, scfg.WriteTimeout)
	}
}

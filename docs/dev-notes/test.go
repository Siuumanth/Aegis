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
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Step 1: parse yaml
	rawConfig, err := config.Load("aegis.yaml")
	if err != nil {
		panic(err)
	}
	cfg := config.BuildRuntimeConfig(rawConfig)
	//  global context to to pass around, specially for async workers
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	globalCtx, cancel := context.WithCancel(ctx)

	// Building router + dependencies
	router, hk, tag := buildRouter(cfg, rawConfig, globalCtx)

	// start server and for each connection, handle it
	//	fmt.Printf("AEGIS listening on %s:%d\n", rawConfig.Server.Host, rawConfig.Server.Port)

	// start server
	if err := serve(rawConfig.Server, router, globalCtx); err != nil {
		fmt.Println("Server error:", err)
	}
	cancel()
	// graceful shutdown, wait for all async workers to finish
	cleanup(hk, tag)

	fmt.Println("AEGIS TCP Server Stopped.")
}

// client = connection from app
func handleConnection(conn net.Conn, r *proxy.Router, globalCtx context.Context, readTimeout *time.Duration, writeTimeout *time.Duration) {
	parser := resp.NewParser(conn)
	// get new conn
	pconn := proxy.NewConn(conn, r, parser, readTimeout, writeTimeout)
	pconn.Handle(globalCtx)
}

// buildRouter wires all core dependencies and returns the request router
func buildRouter(cfg *config.RuntimeConfig, rawConfig *config.Config, globalCtx context.Context) (*proxy.Router, *hotkeys.HotKeyService, *tags.TagService) {
	// wrap circuit breaker over base data layer for resilience
	redisCBClient := redis.NewCBBackend(redis.NewClient(rawConfig.Redis), rawConfig.Redis)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // global context
	defer cancel()
	if err := redisCBClient.Ping(ctx); err != nil {
		log.Fatalf("Redis is down or unreachable\nExiting...")
	}

	// optional subsystems (injected into handler)
	hk := hotkeys.NewHotKeyService(cfg.GlobalConfig, redisCBClient, config.DefaultBufSize)
	tag := tags.NewTagService(cfg.GlobalConfig, redisCBClient, config.DefaultBufSize)

	// start async workers for background processing
	if hk != nil {
		hk.Init(globalCtx, config.DefaultHotKeyWorkers)
	}
	if tag != nil {
		tag.Init(globalCtx, config.DefaultTagWorkers)
	}

	// handler, executes commands + applies cache logic
	h := handler.NewHandler(redisCBClient, hk, tag, rawConfig.Redis.Address)

	p := policy.NewEngine(cfg) // policy engine evaluates yaml rules per request

	// router is the entry point that ties handler + policy together
	router := proxy.NewRouter(cfg, h, p)
	return router, hk, tag
}

func serve(scfg *config.ServerConfig, router *proxy.Router, globalCtx context.Context) error {
	addr := fmt.Sprintf("%s:%d", scfg.Host, scfg.Port)
	// main tcp listen cmd
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic("Init Error" + err.Error())
	}
	defer ln.Close()
	fmt.Println("AEGIS is listening on PORT:", scfg.Port)

	// close listener when context is cancelled, which unblocks Accept()
	go func() {
		<-globalCtx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept() // blocking
		if err != nil {
			select {
			case <-globalCtx.Done():
				return nil // graceful shutdown
			default:
				fmt.Println("Accept error:", err)
				continue
			}
		}

		go handleConnection(conn, router, globalCtx, &scfg.ReadTimeout, &scfg.WriteTimeout)
	}
}

func cleanup(hk *hotkeys.HotKeyService, tag *tags.TagService) {
	// wait - waits till all async workers return
	// stop - closes channel
	if hk != nil {
		hk.Stop()
	}
	if tag != nil {
		tag.Stop()
	}
	if hk != nil {
		hk.Wait()
	}
	if tag != nil {
		tag.Wait()
	}
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// DEV SETTINGS
const (
	TEST_COUNT = 100 // Number of requests for benchmark
	ADDR       = "localhost:6379"
)

func main() {
	ctx := context.Background()

	// mode := "load"
	mode := ""
	// it for interactive

	rdb := redis.NewClient(&redis.Options{
		Addr:         ADDR,
		Protocol:     2,
		WriteTimeout: 200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
	})

	switch mode {
	case "it":
		runREPL(ctx, rdb)
	case "load":
		runBenchmark(ctx, rdb)
	default:
		runBenchmark(ctx, rdb)

		runREPL(ctx, rdb)
	}
}

func runBenchmark(ctx context.Context, rdb *redis.Client) {
	rdb.FlushAll(ctx)
	fmt.Printf("🚀 Running Benchmark: %d\n", TEST_COUNT)

	var total time.Duration
	success := 0

	for i := 0; i < TEST_COUNT; i++ {
		key := fmt.Sprintf("user:%d", rand.Intn(100))
		val := fmt.Sprintf("data-%d", i)
		//ttl := 100 * time.Second

		start := time.Now()

		// SET + GET sequence
		if err := rdb.Do(ctx, "SET", key, val).Err(); err != nil {
			fmt.Printf("SET Error: %v\n", err)
			continue
		}

		if _, err := rdb.Get(ctx, key).Result(); err != nil {
			fmt.Printf("GET Error: %v\n", err)
			continue
		}

		// Skip first request for accurate average (warm-up)
		if i > 0 {
			total += time.Since(start)
			success++
		}
	}

	if success > 0 {
		log.Printf("AVERAGE LATENCY: %v", total/time.Duration(success))
	}
}

func runREPL(ctx context.Context, rdb *redis.Client) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("⌨️  Aegis REPL (%s) - Type 'exit' to quit\n", ADDR)

	for {
		fmt.Print("aegis> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			break
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		args := make([]interface{}, len(parts))
		for i, v := range parts {
			args[i] = v
		}

		start := time.Now()
		res, err := rdb.Do(ctx, args...).Result()

		if err != nil {
			fmt.Printf("(error) %v\n", err)
		} else {
			fmt.Printf("%v (%v)\n", res, time.Since(start))
		}
	}
}

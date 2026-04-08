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
	TEST_COUNT = 1000 // Number of requests for benchmark
	AEGIS_ADDR = "localhost:6380"
	REDIS_ADDR = "localhost:6379"
)

func main() {
	ctx := context.Background()

	// Mode options: "it", "load", "compare", or "" (default)
	//	mode := "compare"
	mode := "tp"

	rdbAegis := redis.NewClient(&redis.Options{
		Addr:         AEGIS_ADDR,
		Protocol:     2,
		WriteTimeout: 200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
	})

	rdbDirect := redis.NewClient(&redis.Options{
		Addr:         REDIS_ADDR,
		Protocol:     2,
		WriteTimeout: 200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
	})

	switch mode {
	case "it":
		runREPL(ctx, rdbAegis)
	case "load":
		runBenchmark(ctx, rdbAegis, "AEGIS")
	case "compare":
		runComparison(ctx, rdbAegis, rdbDirect)
	case "tp":
		runConcurrentBenchmark(ctx, rdbDirect, "DIRECT REDIS", 50)
		runConcurrentBenchmark(ctx, rdbAegis, "AEGIS PROXY", 50)
	default:
		runBenchmark(ctx, rdbAegis, "AEGIS")
		runREPL(ctx, rdbAegis)
	}
}

func runComparison(ctx context.Context, aegis *redis.Client, direct *redis.Client) {
	//fmt.Println("📊 Starting Side-by-Side Comparison")

	// Run Direct Redis first
	redisLatency := runBenchmark(ctx, direct, "DIRECT REDIS (6379)")

	// Run Aegis second
	aegisLatency := runBenchmark(ctx, aegis, "AEGIS PROXY (6380)")

	fmt.Printf("📍 Direct Redis: %v\n", redisLatency)
	fmt.Printf("🛡️  Aegis Proxy:  %v\n", aegisLatency)

	if redisLatency > 0 {
		diff := aegisLatency - redisLatency
		overhead := (float64(diff) / float64(redisLatency)) * 100
		fmt.Printf("⚠️  Overhead:     +%v (%0.2f%%)\n", diff, overhead)
	}
	fmt.Println("------------------------------------")
}

func runBenchmark(ctx context.Context, rdb *redis.Client, label string) time.Duration {
	// rdb.FlushAll(ctx) // Optional: be careful using this in comparison
	fmt.Printf("🚀 Running %s Benchmark: %d reqs\n", label, TEST_COUNT)

	var total time.Duration
	success := 0

	for i := 0; i < TEST_COUNT; i++ {
		key := fmt.Sprintf("user:%d", rand.Intn(100))
		val := fmt.Sprintf("data-%d", i)

		start := time.Now()

		if err := rdb.Do(ctx, "SET", key, val).Err(); err != nil {
			continue
		}

		if _, err := rdb.Get(ctx, key).Result(); err != nil {
			continue
		}

		// Skip warm-up request
		if i > 0 {
			total += time.Since(start)
			success++
		}
	}

	avg := time.Duration(0)
	if success > 0 {
		avg = total / time.Duration(success)
		log.Printf("[%s] AVG LATENCY: %v", label, avg)
	}
	return avg
}

func runREPL(ctx context.Context, rdb *redis.Client) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("⌨️  Aegis REPL (%s) - Type 'exit' to quit\n", AEGIS_ADDR)

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

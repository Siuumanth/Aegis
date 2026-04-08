package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

func runConcurrentBenchmark(ctx context.Context, rdb *redis.Client, label string, workers int) time.Duration {
	fmt.Printf("🚀 Running %s Concurrent Benchmark: %d reqs, %d workers\n", label, TEST_COUNT, workers)

	var (
		total   int64
		success int64
		wg      sync.WaitGroup
	)

	jobs := make(chan int, TEST_COUNT)
	for i := 0; i < TEST_COUNT; i++ {
		jobs <- i
	}
	close(jobs)

	start := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				key := fmt.Sprintf("user:%d", rand.Intn(100))
				val := fmt.Sprintf("data-%d", i)
				if err := rdb.Do(ctx, "SET", key, val).Err(); err != nil {
					continue
				}
				if _, err := rdb.Get(ctx, key).Result(); err != nil {
					continue
				}
				atomic.AddInt64(&total, 1)
				atomic.AddInt64(&success, 1)
			}
		}()
	}
	wg.Wait()

	elapsed := time.Since(start)
	throughput := float64(success) / elapsed.Seconds()
	log.Printf("[%s] %d reqs in %v | %.0f req/s", label, success, elapsed, throughput)
	return elapsed
}

package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// Connect to AEGIS on 6379
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	fmt.Println("Starting continuous Stress Test on Aegis...")

	for i := 0; ; i++ {
		key := fmt.Sprintf("user:%d", rand.Intn(100)) // Rotate through 100 keys
		val := fmt.Sprintf("data-%d", i)

		// 1. SET
		err := rdb.Set(ctx, key, val, 10*time.Second).Err()
		if err != nil {
			fmt.Printf("SET Error: %v\n", err)
		} else {
			fmt.Printf("[%d] SET %s -> %s\n", i, key, val)
		}

		// 2. GET
		getVal, err := rdb.Get(ctx, key).Result()
		if err != nil {
			fmt.Printf("GET Error: %v\n", err)
		} else {
			fmt.Printf("[%d] GET %s -> %s\n", i, key, getVal)
		}

		// Small sleep so you can actually read the logs
		time.Sleep(500 * time.Millisecond)
	}
}

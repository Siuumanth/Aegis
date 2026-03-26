package main

import (
	"context"
	"fmt"
	"log"
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

	fmt.Println("Starting continuous cache Test on Aegis...")
	var total time.Duration

	for i := 0; i < 10; i++ {
		start := time.Now()

		key := fmt.Sprintf("user:%d", rand.Intn(100)) // Rotate through 100 keys
		val := fmt.Sprintf("data-%d", i)

		// 1. SET
		err := rdb.Set(ctx, key, val, 10*time.Second).Err()
		if err != nil {
			fmt.Printf("SET Error: %v\n", err)
		} else {
			//fmt.Printf("[%d] SET %s -> %s\n", i, key, val)
		}

		// 2. GET
		_, err = rdb.Get(ctx, key).Result()
		if err != nil {
			fmt.Printf("GET Error: %v\n", err)
		} else {
			//	fmt.Printf("[%d] GET %s -> %s\n", i, key, getVal)
		}
		end := time.Since(start)
		log.Println(end)
		total += end
		// Small sleep so you can actually read the logs
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("AVERAGE TIME IS ", total/time.Duration(10))

}

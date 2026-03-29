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
		Addr:         "localhost:6379",
		Protocol:     2,
		WriteTimeout: 100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
	})

	fmt.Println("Starting continuous cache Test on Aegis...")
	var total time.Duration
	count := 100
	//	exp := 20

	for i := 0; i < count; i++ {
		start := time.Now()

		key := fmt.Sprintf("user:%d", rand.Intn(100))
		val := fmt.Sprintf("data-%d", i)

		// SET with expiry (EX 10 seconds)
		//err := rdb.Do(ctx, "SET", key, val, "EX", exp).Err()
		err := rdb.Do(ctx, "SET", key, val).Err()
		if err != nil {
			fmt.Printf("SET Error: %v\n", err)
			continue
		}

		// GET
		_, err = rdb.Do(ctx, "GET", key).Result()
		if err != nil {
			fmt.Printf("GET Error: %v\n", err)
		}

		end := time.Since(start)
		if i == 0 {
			continue
		}
		//log.Println(end)
		// if firs time then skip

		total += end

		//time.Sleep(500 * time.Millisecond)
	}

	log.Println("AVERAGE TIME IS ", total/time.Duration(count))
}

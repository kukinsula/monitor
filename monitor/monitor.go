package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kukinsula/monitor/mq"
)

func main() {
	fmt.Println("Connecting...")

	service, err := mq.NewService(":6379")
	if err != nil {
		fmt.Printf("Redis connection failed: %v", err)
		return
	}

	defer service.Close()

	fmt.Println("Successfully connected!")

	for {
		time.Sleep(time.Duration(random(500, 1000)) * time.Millisecond)

		metrics := &mq.Metrics{
			CPU: map[string]interface{}{
				"processors": []int64{
					random(0, 100),
					random(0, 100),
					random(0, 100),
					random(0, 100),
				},
			},

			RAM: map[string]interface{}{
				"free": random(1000000, 16000000),
				"used": random(1000000, 16000000),
			},

			// "ROM": map[string]interface{}{},
			// "NET": map[string]interface{}{},
		}

		err = service.PublishMetrics(metrics)
		if err != nil {
			fmt.Printf("PUB Metrics failed: %v\n", metrics, err)
			return
		}
	}

	fmt.Println("Done")
}

func random(min, max int) int64 {
	return rand.Int63n(int64(max-min)) + int64(min)
}

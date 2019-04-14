package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kukinsula/monitor/mq"
)

func main() {
	fmt.Print("MONITOR: connecting...")

	service := mq.NewService(":6379")
	defer service.Close()

	fmt.Println("\nMONITOR: successfully connected!")

	_, err := service.Ping()
	if err != nil {
		fmt.Printf("MONITOR: Error cannot 'PING': %v", err)
	}

	for {
		time.Sleep(time.Duration(random(100, 500)) * time.Millisecond)

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
			fmt.Printf("MONITOR: PUB Metrics failed: %v\n", metrics, err)
			return
		}
	}

	fmt.Println("MONITOR: done!")
}

func random(min, max int) int64 {
	return rand.Int63n(int64(max-min)) + int64(min)
}

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"

	"github.com/kukinsula/monitor/mq"
)

func main() {
	fmt.Println("Connecting...")

	service, err := mq.NewService(":6379")
	if err != nil {
		fmt.Printf("Redis connection failed %v", err)
		return
	}

	fmt.Println("Successfully connected!")

	quit := make(chan os.Signal, 1)
	cleanup := make(chan struct{}, 1)

	signal.Notify(quit, os.Interrupt)

	defer close(quit)
	defer close(cleanup)

	api := NewAPI(service)

	go api.Run()

	go func() {
		<-quit

		fmt.Println("Quiting...")

		err := api.Shutdown()
		if err != nil {
			fmt.Println("API Shutdown failed:", err)
		}

		service.Close()
		if err != nil {
			fmt.Println("API Shutdown failed:", err)
		}

		cleanup <- struct{}{}
	}()

	<-cleanup

	fmt.Println("Done")
}

func random(min, max int) int64 {
	return rand.Int63n(int64(max-min)) + int64(min)
}
